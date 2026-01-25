# 追踪日志使用指南

## 更新说明 (2026-01-25)

为了让追踪日志更加明显和易于识别，我们对日志格式进行了优化。

---

## 新的日志格式

所有追踪日志现在都有 **`[TRACE]`** 前缀和 **`trace_type`** 字段，方便过滤和识别。

### 1. 请求开始日志

```log
time="2026-01-25 00:26:55" level=INFO msg="[TRACE] request_started"
  trace_type="request_lifecycle"       👈 追踪类型标识
  handler="gemini"                     👈 处理器类型 (gemini/anthropic/openai)
  request_id="2b212cac-529f-41a3-a240-d8bb0525a7a4"
  endpoint="/v1beta/models/gemini-3-flash-preview:countTokens"
  method="POST"
  ip="3.145.142.42"
  start_time="2026-01-25T00:26:55.880+08:00"
```

### 2. 请求完成日志

```log
time="2026-01-25 00:27:12" level=INFO msg="[TRACE] request_completed"
  trace_type="request_lifecycle"       👈 追踪类型标识
  handler="gemini"                     👈 处理器类型
  request_id="2b212cac-529f-41a3-a240-d8bb0525a7a4"
  user_id=1                            👈 用户ID
  model="gemini-3-flash-preview"       👈 模型名称
  account_id=5                         👈 上游账户ID
  status=200                           👈 HTTP状态码
  cost_usd=0                           👈 费用 (当前为占位符)
  duration_ms=16305                    👈 总耗时
  end_time="2026-01-25T00:27:12.186+08:00"
```

### 3. 性能分析日志 (现有系统,未修改)

```log
time="2026-01-25 00:27:12" level=INFO msg="" msg=request_completed
  request_id="2b212cac-529f-41a3-a240-d8bb0525a7a4"
  user_id=1
  model="gemini-3-flash-preview"
  total_ms=16311                       👈 总耗时
  wait_ms=0                            👈 并发等待耗时
  upstream_ms=16301                    👈 上游API耗时
  internal_ms=10                       👈 内部处理耗时
  account_id=5
  platform="gemini"
```

---

## 支持的 API 类型

追踪日志覆盖以下所有 API 处理器：

| handler 值 | API 类型 | 端点示例 |
|-----------|---------|---------|
| `anthropic` | Claude/Anthropic API | `/v1/messages` |
| `gemini` | Gemini API | `/v1beta/models/{model}:generateContent` |
| `openai` | OpenAI Compatible API | `/openai/v1/responses` |

---

## 快速过滤日志

### 只看追踪日志

```bash
# 方法1: 通过 [TRACE] 前缀过滤
tail -f logs/app.log | grep "\[TRACE\]"

# 方法2: 通过 trace_type 字段过滤
tail -f logs/app.log | grep "trace_type="
```

### 只看请求开始日志

```bash
tail -f logs/app.log | grep "\[TRACE\] request_started"
```

### 只看请求完成日志

```bash
tail -f logs/app.log | grep "\[TRACE\] request_completed"
```

### 按 handler 类型过滤

```bash
# 只看 Gemini API 的追踪日志
tail -f logs/app.log | grep "\[TRACE\]" | grep "handler=gemini"

# 只看 Claude API 的追踪日志
tail -f logs/app.log | grep "\[TRACE\]" | grep "handler=anthropic"
```

### 按 request_id 追踪完整请求

```bash
# 替换为实际的 request_id
REQUEST_ID="2b212cac-529f-41a3-a240-d8bb0525a7a4"
tail -f logs/app.log | grep "$REQUEST_ID"
```

这会显示该请求的所有日志：
- `[TRACE] request_started` - 请求开始
- `[TRACE] request_completed` - 业务追踪
- `request_completed` (无前缀) - 性能分析

---

## 日志对比表

| 特征 | 追踪日志 (新) | 性能日志 (旧) |
|-----|-------------|--------------|
| **msg 前缀** | `[TRACE]` | 无 |
| **trace_type 字段** | ✅ 有 | ❌ 无 |
| **handler 字段** | ✅ 有 | ❌ 无 |
| **start_time 字段** | ✅ 有 (开始日志) | ❌ 无 |
| **end_time 字段** | ✅ 有 (完成日志) | ❌ 无 |
| **duration_ms** | ✅ 有 | ✅ 有 (total_ms) |
| **wait_ms** | ❌ 无 | ✅ 有 |
| **upstream_ms** | ❌ 无 | ✅ 有 |
| **internal_ms** | ❌ 无 | ✅ 有 |

**互补关系**:
- 追踪日志: 关注**请求生命周期**和**业务信息**
- 性能日志: 关注**性能分析**和**阶段耗时**

---

## 完整示例

一个完整的 Gemini API 请求会产生以下日志序列：

```log
# 1️⃣ 请求开始
time="..." level=INFO msg="[TRACE] request_started"
  trace_type="request_lifecycle"
  handler="gemini"
  request_id="xxx"
  endpoint="/v1beta/models/gemini-3-flash-preview:streamGenerateContent"
  method="POST"
  ip="3.145.142.42"
  start_time="2026-01-25T00:27:12.886+08:00"

# 2️⃣ 账户调度日志 (DEBUG 级别)
time="..." level=DEBUG msg=account_scheduling_starting
time="..." level=DEBUG msg=account_scheduling_account_detail

# 3️⃣ Gemini 上游请求日志
time="..." level=INFO msg="[GeminiNative] account=..."
time="..." level=INFO msg="[GeminiAPI] ========== Streaming Response Headers =========="

# 4️⃣ 追踪日志 - 请求完成
time="..." level=INFO msg="[TRACE] request_completed"
  trace_type="request_lifecycle"
  handler="gemini"
  request_id="xxx"
  user_id=1
  model="gemini-3-flash-preview"
  account_id=5
  status=200
  cost_usd=0
  duration_ms=4128
  end_time="2026-01-25T00:27:17.014+08:00"

# 5️⃣ GIN 访问日志
time="..." level=INFO msg="[GIN] 2026/01/25 - 00:27:17 | 200 | 4.131s | ... | POST ..."

# 6️⃣ 性能分析日志 (现有系统)
time="..." level=INFO msg="" msg=request_completed
  request_id="xxx"
  user_id=1
  model="gemini-3-flash-preview"
  status=200
  total_ms=4131
  wait_ms=0
  upstream_ms=4125
  internal_ms=6
  account_id=5
  platform="gemini"
  stream=true
```

---

## Claude API 测试

如果你想看 Claude API 的追踪日志，可以发送测试请求：

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "x-api-key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 100,
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

然后查看日志：

```bash
tail -f logs/app.log | grep "\[TRACE\]" | grep "handler=anthropic"
```

你会看到类似这样的输出：

```log
msg="[TRACE] request_started"
  trace_type="request_lifecycle"
  handler="anthropic"        👈 Claude API
  endpoint="/v1/messages"
  ...

msg="[TRACE] request_completed"
  trace_type="request_lifecycle"
  handler="anthropic"        👈 Claude API
  model="claude-3-5-sonnet-20241022"
  ...
```

---

## 配置控制

在 `config.yaml` 中控制追踪日志的开关：

```yaml
logging:
  request_tracing:
    enabled: true
    detailed: true
    verbose_logging: true   # ← 控制 [TRACE] 日志的开关
    sample_rate: 1.0        # 1.0 = 100% 请求都记录
```

### 场景化配置

**开发环境** (查看所有请求):
```yaml
verbose_logging: true
sample_rate: 1.0
```

**生产环境** (降低日志量):
```yaml
verbose_logging: true
sample_rate: 0.1   # 只记录 10% 的请求
```

**性能敏感场景** (完全关闭):
```yaml
verbose_logging: false
```

---

## 常见问题

### Q1: 为什么有些请求没有 `[TRACE]` 日志？

**A**: 只有通过以下 handler 的请求才会有追踪日志：
- `gateway_handler.Messages()` - Claude/Anthropic API
- `gateway_handler.GeminiV1BetaModels()` - Gemini API

其他路由 (如 `/api/v1/auth/me`, `/api/v1/subscriptions/active`) 不会有 `[TRACE]` 日志。

### Q2: `[TRACE]` 日志和无前缀的 `request_completed` 有什么区别？

**A**:
- `[TRACE] request_completed` - **新增的追踪日志**，关注业务信息 (user_id, model, handler)
- `msg="" msg=request_completed` - **现有的性能日志**，关注性能分析 (wait_ms, upstream_ms)

两者互补，各有用途。

### Q3: `cost_usd=0` 是正常的吗？

**A**: 是的。费用计算在 `RecordUsage` 中异步完成并记录到数据库，不在日志中实时显示。这个字段是占位符，未来可以优化为显示实际费用。

### Q4: 如何统计每个用户的请求量？

**A**:
```bash
# 统计每个用户的请求次数
grep "\[TRACE\] request_completed" logs/app.log | \
  grep -o "user_id=[0-9]*" | \
  sort | uniq -c
```

### Q5: 如何找出最慢的请求？

**A**:
```bash
# 找出耗时超过 10 秒的请求
grep "\[TRACE\] request_completed" logs/app.log | \
  awk '$NF > 10000 {print}' FS='duration_ms=' | \
  sort -t= -k2 -n
```

---

## 总结

✅ 所有追踪日志现在都有明显的 **`[TRACE]`** 前缀
✅ 通过 `trace_type` 字段区分日志类型
✅ 通过 `handler` 字段区分 API 类型 (gemini/anthropic/openai)
✅ 支持灵活的 grep 过滤
✅ 与现有性能日志互补，不冲突

使用 `grep "\[TRACE\]"` 即可快速查看所有追踪日志！
