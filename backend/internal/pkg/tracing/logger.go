package tracing

import (
	"context"
	"log/slog"
)

// LogRequestMetrics 异步记录请求追踪指标
// 在 goroutine 中调用，不阻塞请求处理
func LogRequestMetrics(tracer *RequestTracer, statusCode int) {
	if tracer == nil {
		return
	}

	total, wait, upstream, internal := tracer.CalculateDurations()

	// 构建日志字段
	attrs := []slog.Attr{
		slog.String("msg", "request_completed"),
		slog.String("request_id", tracer.RequestID),
		slog.Int64("user_id", tracer.UserID),
		slog.String("model", tracer.Model),
		slog.Int("status", statusCode),
		slog.Int64("total_ms", total),
		slog.Int64("wait_ms", wait),
		slog.Int64("upstream_ms", upstream),
		slog.Int64("internal_ms", internal),
	}

	// 添加可选字段
	if tracer.AccountID != nil {
		attrs = append(attrs, slog.Int64("account_id", *tracer.AccountID))
	}
	if tracer.Platform != "" {
		attrs = append(attrs, slog.String("platform", tracer.Platform))
	}
	if tracer.Stream {
		attrs = append(attrs, slog.Bool("stream", tracer.Stream))
	}

	// 详细模式：记录各阶段时间戳
	if tracer.detailed {
		waitStart := tracer.GetStage(StageWaitStart)
		waitEnd := tracer.GetStage(StageWaitEnd)
		upstreamStart := tracer.GetStage(StageUpstreamStart)
		upstreamEnd := tracer.GetStage(StageUpstreamEnd)

		if !waitStart.IsZero() {
			attrs = append(attrs, slog.Time("wait_start", waitStart))
		}
		if !waitEnd.IsZero() {
			attrs = append(attrs, slog.Time("wait_end", waitEnd))
		}
		if !upstreamStart.IsZero() {
			attrs = append(attrs, slog.Time("upstream_start", upstreamStart))
		}
		if !upstreamEnd.IsZero() {
			attrs = append(attrs, slog.Time("upstream_end", upstreamEnd))
		}
	}

	// 输出 INFO 级别日志
	slog.LogAttrs(context.Background(), slog.LevelInfo, "", attrs...)
}
