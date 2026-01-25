package middleware

import (
	"math/rand"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tracing"
	"github.com/gin-gonic/gin"
)

// RequestTracer 中间件：追踪请求耗时
// 根据配置启用/禁用追踪，支持采样率控制
func RequestTracer(cfg *config.Config) gin.HandlerFunc {
	// 如果未配置或未启用，返回空中间件
	if cfg == nil || !cfg.Logging.RequestTracing.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	sampleRate := cfg.Logging.RequestTracing.SampleRate
	detailed := cfg.Logging.RequestTracing.Detailed

	// 确保采样率在 0.0-1.0 范围内
	if sampleRate < 0.0 {
		sampleRate = 0.0
	}
	if sampleRate > 1.0 {
		sampleRate = 1.0
	}

	return func(c *gin.Context) {
		// 采样判断：如果采样率 < 1.0，随机决定是否追踪
		if sampleRate < 1.0 && rand.Float64() > sampleRate {
			// 不追踪这个请求
			c.Next()
			return
		}

		// 获取 request ID (由 RequestID 中间件设置)
		requestID := tracing.GetRequestID(c)
		if requestID == "" {
			// 如果没有 request ID，可能是中间件顺序错误，跳过追踪
			c.Next()
			return
		}

		// 创建追踪器
		tracer := tracing.NewRequestTracer(requestID, detailed)
		tracing.SetTracer(c, tracer)

		// 请求处理完成后记录日志并释放追踪器
		defer func() {
			statusCode := c.Writer.Status()
			// 异步记录日志，不阻塞响应
			go func() {
				tracing.LogRequestMetrics(tracer, statusCode)
				tracer.Release()
			}()
		}()

		c.Next()
	}
}
