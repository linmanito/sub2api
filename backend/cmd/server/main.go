package main

//go:generate go run github.com/google/wire/cmd/wire

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/setup"
	"github.com/Wei-Shaw/sub2api/internal/web"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

//go:embed VERSION
var embeddedVersion string

// Build-time variables (can be set by ldflags)
var (
	Version   = ""
	Commit    = "unknown"
	Date      = "unknown"
	BuildType = "source" // "source" for manual builds, "release" for CI builds (set by ldflags)
)

func init() {
	// Read version from embedded VERSION file
	Version = strings.TrimSpace(embeddedVersion)
	if Version == "" {
		Version = "0.0.0-dev"
	}
}

// initLogger configures the default slog handler based on gin.Mode().
// Logs are written to daily rotated files in the logs directory.
// In non-release mode, Debug level logs are enabled.
func initLogger() {
	var level slog.Level
	if gin.Mode() == gin.ReleaseMode {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}

	// 确保日志目录存在
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v, falling back to stderr", err)
		// 回退到 stderr
		handler := NewCustomHandler(os.Stderr, level)
		slog.SetDefault(slog.New(handler))
		return
	}

	// 配置 file-rotatelogs 进行按天轮换
	// 工作原理（软链接方式）:
	// 1. 真实日志文件: sub2api-2026-01-24.log（按日期命名）
	// 2. 软链接: sub2api.log -> sub2api-2026-01-24.log（始终指向当天日志）
	// 3. 每天 00:00，自动创建新日期的日志文件，并更新软链接
	// 4. 始终可以用 tail -f sub2api.log 查看最新日志
	logPath := filepath.Join(logDir, "sub2api.log")
	logFile, err := rotatelogs.New(
		filepath.Join(logDir, "sub2api-%Y-%m-%d.log"), // 真实文件名格式（带日期）
		rotatelogs.WithLinkName(logPath),               // 软链接文件名（固定不变）
		rotatelogs.WithRotationTime(24*time.Hour),      // 每24小时轮换一次
		rotatelogs.WithMaxAge(90*24*time.Hour),         // 保留90天
	)
	if err != nil {
		log.Printf("Failed to create rotating log file: %v, falling back to stderr", err)
		handler := NewCustomHandler(os.Stderr, level)
		slog.SetDefault(slog.New(handler))
		return
	}

	// 同时输出到文件和控制台(便于 systemd 也能看到日志)
	multiWriter := io.MultiWriter(logFile, os.Stderr)

	// 使用自定义 Handler 以支持自定义时间格式
	handler := NewCustomHandler(multiWriter, level)
	slog.SetDefault(slog.New(handler))

	log.Printf("Logger initialized with daily rotation in %s directory", logDir)
}

func main() {
	// Initialize slog logger based on gin mode
	initLogger()

	// Parse command line flags
	setupMode := flag.Bool("setup", false, "Run setup wizard in CLI mode")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		log.Printf("Sub2API %s (commit: %s, built: %s)\n", Version, Commit, Date)
		return
	}

	// CLI setup mode
	if *setupMode {
		if err := setup.RunCLI(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		return
	}

	// Check if setup is needed
	if setup.NeedsSetup() {
		// Check if auto-setup is enabled (for Docker deployment)
		if setup.AutoSetupEnabled() {
			log.Println("Auto setup mode enabled...")
			if err := setup.AutoSetupFromEnv(); err != nil {
				log.Fatalf("Auto setup failed: %v", err)
			}
			// Continue to main server after auto-setup
		} else {
			log.Println("First run detected, starting setup wizard...")
			runSetupServer()
			return
		}
	}

	// Normal server mode
	runMainServer()
}

func runSetupServer() {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(config.CORSConfig{}))
	r.Use(middleware.SecurityHeaders(config.CSPConfig{Enabled: true, Policy: config.DefaultCSPPolicy}))

	// Register setup routes
	setup.RegisterRoutes(r)

	// Serve embedded frontend if available
	if web.HasEmbeddedFrontend() {
		r.Use(web.ServeEmbeddedFrontend())
	}

	// Get server address from config.yaml or environment variables (SERVER_HOST, SERVER_PORT)
	// This allows users to run setup on a different address if needed
	addr := config.GetServerAddress()
	log.Printf("Setup wizard available at http://%s", addr)
	log.Println("Complete the setup wizard to configure Sub2API")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start setup server: %v", err)
	}
}

func runMainServer() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if cfg.RunMode == config.RunModeSimple {
		log.Println("⚠️  WARNING: Running in SIMPLE mode - billing and quota checks are DISABLED")
	}

	buildInfo := handler.BuildInfo{
		Version:   Version,
		BuildType: BuildType,
	}

	app, err := initializeApplication(buildInfo)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Cleanup()

	// 启动服务器
	go func() {
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on %s", app.Server.Addr)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
