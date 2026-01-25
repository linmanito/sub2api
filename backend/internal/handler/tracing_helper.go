package handler

import (
	"log/slog"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tracing"
	"github.com/gin-gonic/gin"
)

// TracingHelper 请求追踪辅助工具
type TracingHelper struct {
	cfg *config.Config
}

// NewTracingHelper 创建追踪辅助工具
func NewTracingHelper(cfg *config.Config) *TracingHelper {
	return &TracingHelper{cfg: cfg}
}

// LogRequestStart 记录请求开始日志
// 返回请求开始时间，用于计算总耗时
func (h *TracingHelper) LogRequestStart(c *gin.Context, handler string) time.Time {
	startTime := time.Now()

	// 只有在启用 verbose_logging 时才记录详细日志
	if !h.cfg.Logging.RequestTracing.VerboseLogging {
		return startTime
	}

	requestID := tracing.GetRequestID(c)

	slog.Info("[TRACE] request_started",
		"trace_type", "request_lifecycle",
		"handler", handler,
		"request_id", requestID,
		"endpoint", c.Request.URL.Path,
		"method", c.Request.Method,
		"ip", c.ClientIP(),
		"start_time", startTime)

	return startTime
}

// LogRequestCompleted 记录请求完成日志
func (h *TracingHelper) LogRequestCompleted(
	c *gin.Context,
	handler string,
	startTime time.Time,
	userID int64,
	model string,
	accountID int64,
	statusCode int,
	costUSD float64,
) {
	// 只有在启用 verbose_logging 时才记录详细日志
	if !h.cfg.Logging.RequestTracing.VerboseLogging {
		return
	}

	requestID := tracing.GetRequestID(c)

	slog.Info("[TRACE] request_completed",
		"trace_type", "request_lifecycle",
		"handler", handler,
		"request_id", requestID,
		"user_id", userID,
		"model", model,
		"account_id", accountID,
		"status", statusCode,
		"cost_usd", costUSD,
		"duration_ms", time.Since(startTime).Milliseconds(),
		"end_time", time.Now())
}

// LogStage 记录请求处理阶段
func (h *TracingHelper) LogStage(c *gin.Context, stage string, details map[string]any) {
	// 只有在启用 verbose_logging 时才记录详细日志
	if !h.cfg.Logging.RequestTracing.VerboseLogging {
		return
	}

	requestID := tracing.GetRequestID(c)

	attrs := []any{
		"trace_type", "request_stage",
		"request_id", requestID,
		"stage", stage,
		"timestamp", time.Now(),
	}

	// 添加额外的详细信息
	for k, v := range details {
		attrs = append(attrs, k, v)
	}

	slog.Debug("[TRACE] request_stage", attrs...)
}

// SetupTracer 设置 tracer 的基础信息
func (h *TracingHelper) SetupTracer(c *gin.Context, userID int64, model string) {
	if tracer := tracing.GetTracer(c); tracer != nil {
		tracer.SetUserID(userID)
		tracer.SetModel(model)
	}
}

// SetTracerAccount 设置 tracer 的账户信息
func (h *TracingHelper) SetTracerAccount(c *gin.Context, accountID int64, platform string) {
	if tracer := tracing.GetTracer(c); tracer != nil {
		tracer.SetAccountID(accountID)
		tracer.SetPlatform(platform)
		tracer.MarkAccountSelect()
	}
}

// MarkWaitStart 标记并发等待开始
func (h *TracingHelper) MarkWaitStart(c *gin.Context) {
	if tracer := tracing.GetTracer(c); tracer != nil {
		tracer.MarkWaitStart()
	}
}

// MarkWaitEnd 标记并发等待结束
func (h *TracingHelper) MarkWaitEnd(c *gin.Context) {
	if tracer := tracing.GetTracer(c); tracer != nil {
		tracer.MarkWaitEnd()
	}
}

// MarkUpstreamStart 标记上游请求开始
func (h *TracingHelper) MarkUpstreamStart(c *gin.Context) {
	if tracer := tracing.GetTracer(c); tracer != nil {
		tracer.MarkUpstreamStart()
	}
}

// MarkUpstreamEnd 标记上游请求结束
func (h *TracingHelper) MarkUpstreamEnd(c *gin.Context, stream bool) {
	if tracer := tracing.GetTracer(c); tracer != nil {
		tracer.MarkUpstreamEnd()
		tracer.SetStream(stream)
	}
}
