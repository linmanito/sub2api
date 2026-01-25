package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/tracing"
	"github.com/gin-gonic/gin"
)

// RequestID 中间件：为每个请求生成唯一 ID
// 该 ID 会通过 X-Request-ID 响应头返回给客户端
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成并设置 request ID
		requestID := tracing.SetRequestID(c)

		// 添加到响应头，便于客户端追踪
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
