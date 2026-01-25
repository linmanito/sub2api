package tracing

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StageType 追踪阶段类型
type StageType int

const (
	StageWaitStart StageType = iota // 并发等待开始
	StageWaitEnd                     // 并发等待结束
	StageUpstreamStart               // 上游请求开始
	StageUpstreamEnd                 // 上游请求结束
	StageAccountSelect               // 账号选择完成
)

// RequestTracer 请求追踪器
type RequestTracer struct {
	RequestID string
	StartTime time.Time
	UserID    int64
	Model     string
	AccountID *int64
	Platform  string
	Stream    bool

	// 阶段时间戳 (使用固定数组避免堆分配)
	stages    [5]time.Time
	stageIdx  int
	stageLock sync.Mutex

	// 详细追踪模式下的额外字段
	detailed bool
}

// tracerPool 对象池，减少 GC 压力
var tracerPool = sync.Pool{
	New: func() interface{} {
		return &RequestTracer{}
	},
}

// NewRequestTracer 创建新的请求追踪器
func NewRequestTracer(requestID string, detailed bool) *RequestTracer {
	tracer := tracerPool.Get().(*RequestTracer)
	tracer.RequestID = requestID
	tracer.StartTime = time.Now()
	tracer.detailed = detailed
	tracer.stageIdx = 0
	tracer.UserID = 0
	tracer.Model = ""
	tracer.AccountID = nil
	tracer.Platform = ""
	tracer.Stream = false
	// 清空时间戳数组
	for i := range tracer.stages {
		tracer.stages[i] = time.Time{}
	}
	return tracer
}

// Release 释放追踪器回对象池
func (t *RequestTracer) Release() {
	tracerPool.Put(t)
}

// MarkStage 标记阶段时间点
func (t *RequestTracer) MarkStage(stage StageType) {
	t.stageLock.Lock()
	defer t.stageLock.Unlock()

	idx := int(stage)
	if idx < len(t.stages) {
		t.stages[idx] = time.Now()
	}
}

// MarkWaitStart 标记并发等待开始
func (t *RequestTracer) MarkWaitStart() {
	t.MarkStage(StageWaitStart)
}

// MarkWaitEnd 标记并发等待结束
func (t *RequestTracer) MarkWaitEnd() {
	t.MarkStage(StageWaitEnd)
}

// MarkUpstreamStart 标记上游请求开始
func (t *RequestTracer) MarkUpstreamStart() {
	t.MarkStage(StageUpstreamStart)
}

// MarkUpstreamEnd 标记上游请求结束
func (t *RequestTracer) MarkUpstreamEnd() {
	t.MarkStage(StageUpstreamEnd)
}

// MarkAccountSelect 标记账号选择完成
func (t *RequestTracer) MarkAccountSelect() {
	t.MarkStage(StageAccountSelect)
}

// SetUserID 设置用户ID
func (t *RequestTracer) SetUserID(userID int64) {
	t.UserID = userID
}

// SetModel 设置模型
func (t *RequestTracer) SetModel(model string) {
	t.Model = model
}

// SetAccountID 设置账号ID
func (t *RequestTracer) SetAccountID(accountID int64) {
	t.AccountID = &accountID
}

// SetPlatform 设置平台
func (t *RequestTracer) SetPlatform(platform string) {
	t.Platform = platform
}

// SetStream 设置是否流式
func (t *RequestTracer) SetStream(stream bool) {
	t.Stream = stream
}

// GetStage 获取指定阶段的时间戳
func (t *RequestTracer) GetStage(stage StageType) time.Time {
	t.stageLock.Lock()
	defer t.stageLock.Unlock()

	idx := int(stage)
	if idx < len(t.stages) {
		return t.stages[idx]
	}
	return time.Time{}
}

// CalculateDurations 计算各阶段耗时 (毫秒)
func (t *RequestTracer) CalculateDurations() (total, wait, upstream, internal int64) {
	now := time.Now()
	total = now.Sub(t.StartTime).Milliseconds()

	// 并发等待耗时
	waitStart := t.GetStage(StageWaitStart)
	waitEnd := t.GetStage(StageWaitEnd)
	if !waitStart.IsZero() && !waitEnd.IsZero() {
		wait = waitEnd.Sub(waitStart).Milliseconds()
	}

	// 上游请求耗时
	upstreamStart := t.GetStage(StageUpstreamStart)
	upstreamEnd := t.GetStage(StageUpstreamEnd)
	if !upstreamStart.IsZero() && !upstreamEnd.IsZero() {
		upstream = upstreamEnd.Sub(upstreamStart).Milliseconds()
	}

	// 内部处理耗时 = 总耗时 - 等待耗时 - 上游耗时
	internal = total - wait - upstream
	if internal < 0 {
		internal = 0
	}

	return
}

// Context keys
const (
	contextKeyRequestID = "request_id"
	contextKeyTracer    = "request_tracer"
)

// SetRequestID 设置 request ID 到 context
func SetRequestID(c *gin.Context) string {
	requestID := uuid.NewString()
	c.Set(contextKeyRequestID, requestID)
	return requestID
}

// GetRequestID 从 context 获取 request ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(contextKeyRequestID); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// SetTracer 设置 tracer 到 context
func SetTracer(c *gin.Context, tracer *RequestTracer) {
	c.Set(contextKeyTracer, tracer)
}

// GetTracer 从 context 获取 tracer
func GetTracer(c *gin.Context) *RequestTracer {
	if tracer, exists := c.Get(contextKeyTracer); exists {
		if t, ok := tracer.(*RequestTracer); ok {
			return t
		}
	}
	return nil
}
