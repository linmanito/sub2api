package main

import (
	"io"
	"log/slog"
)

// newCustomHandlerOptions 创建自定义 HandlerOptions,支持自定义时间格式
func newCustomHandlerOptions(level slog.Level) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式为 yyyy-MM-dd HH:mm:ss
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, a.Value.Time().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	}
}

// NewCustomHandler 创建自定义 Handler
func NewCustomHandler(w io.Writer, level slog.Level) slog.Handler {
	opts := newCustomHandlerOptions(level)
	return slog.NewTextHandler(w, opts)
}
