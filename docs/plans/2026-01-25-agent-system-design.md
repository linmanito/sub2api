# 代理商系统设计方案

> 版本：v1.0
> 日期：2026-01-25
> 状态：待评审

## 1. 概述

### 1.1 背景

Sub2API 作为 AI API 网关平台，需要支持代理商模式，允许代理商销售平台的订阅/账号服务，扩大用户覆盖面。

### 1.2 目标

- 支持代理商以批发价采购额度，自主定价销售给下游用户
- 提供代理商独立管理后台，管理下游用户和查看收益
- 对平台管理员透明，可监控所有代理商和用户

### 1.3 核心决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 收益模式 | 差价模式 | 代理商批发买入、加价卖出，简单直接 |
| 代理层级 | 单层代理 | 降低系统复杂度，满足 90% 场景 |
| 账号池维护 | 平台统一维护 | 代理商是纯销售角色 |
| 下游用户管理 | 平台托管 | 下游用户在平台注册，归属代理商 |
| 配额单位 | 美元金额 | 统一不同模型价格差异 |
| 额度存储 | 复用 balance 字段 | 减少字段冗余，语义统一 |

---

## 2. 角色与关系

### 2.1 角色定义

```
┌─────────────────────────────────────────────────────────────┐
│                      平台管理员 (admin)                       │
│  - 管理所有用户、代理商                                         │
│  - 给代理商分配额度和可用分组                                    │
│  - 可将普通用户分配给代理商                                      │
└─────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┴───────────────────┐
          ▼                                       ▼
┌─────────────────────┐               ┌─────────────────────┐
│   代理商 (agent)     │               │ 平台直属用户 (user)  │
│  - 拥有可分配额度     │               │  - agent_id = null  │
│  - 管理下游用户       │               │  - 直接由平台服务     │
│  - 自定义下游倍率     │               └─────────────────────┘
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│  下游用户 (user)     │
│  - agent_id = 代理商 │
│  - 由代理商管理       │
└─────────────────────┘
```

### 2.2 数据隔离原则

- 代理商只能看到和操作 `agent_id = 自己ID` 的用户
- 代理商只能分配自己 `allowed_groups` 范围内的分组
- 下游用户消费扣自己的余额，不直接扣代理商

---

## 3. 数据模型变更

### 3.1 users 表新增字段

```sql
ALTER TABLE users ADD COLUMN agent_id BIGINT REFERENCES users(id);
ALTER TABLE users ADD COLUMN agent_rate DECIMAL(10,4);

COMMENT ON COLUMN users.agent_id IS '所属代理商ID，NULL表示平台直属用户';
COMMENT ON COLUMN users.agent_rate IS '代理商给此用户设置的计费倍率，NULL表示使用分组默认倍率';

CREATE INDEX idx_users_agent_id ON users(agent_id) WHERE agent_id IS NOT NULL;
```

### 3.2 新增角色常量

```go
// backend/internal/service/domain_constants.go
const (
    RoleAdmin = "admin"
    RoleAgent = "agent"  // 新增
    RoleUser  = "user"
)
```

### 3.3 新增 agent_invite_codes 表

```sql
CREATE TABLE agent_invite_codes (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        BIGINT NOT NULL REFERENCES users(id),
    code            VARCHAR(32) NOT NULL UNIQUE,
    max_uses        INT NOT NULL DEFAULT 0,        -- 0 表示无限制
    used_count      INT NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- active / disabled
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_invite_codes_agent_id ON agent_invite_codes(agent_id);
CREATE INDEX idx_agent_invite_codes_code ON agent_invite_codes(code) WHERE status = 'active';

COMMENT ON TABLE agent_invite_codes IS '代理商邀请码表';
```

---

## 4. 代理商后台功能

### 4.1 路由结构

```
/agent                     # 代理商后台首页（数据概览）
/agent/users               # 下游用户管理
/agent/users/:id           # 用户详情
/agent/invite-codes        # 邀请码管理
/agent/usage               # 用量统计
/agent/settings            # 代理商设置
```

### 4.2 功能模块

#### 4.2.1 数据概览 `/agent`

| 数据项 | 说明 |
|--------|------|
| 我的余额 | 可分配给下游用户的额度 |
| 下游用户数 | 总数 / 活跃数 |
| 消费统计 | 今日 / 本周 / 本月消费金额 |
| 最近用户 | 最近注册的下游用户列表 |

#### 4.2.2 下游用户管理 `/agent/users`

**列表功能：**
- 搜索（邮箱、用户名）
- 筛选（状态、注册时间）
- 分页

**用户操作：**

| 操作 | 说明 |
|------|------|
| 创建用户 | 邮箱、初始密码、初始余额（可选） |
| 充值 | 从代理商余额划转到用户余额 |
| 扣款 | 从用户余额退回代理商余额 |
| 设置倍率 | 自定义用户计费倍率，低于成本价时弹出警告 |
| 分配订阅 | 给用户分配 subscription 类型分组的订阅 |
| 管理 API Key | 查看、创建、删除用户的 API Key |
| 启用/禁用 | 控制用户状态 |
| 重置密码 | 重置用户登录密码 |

#### 4.2.3 邀请码管理 `/agent/invite-codes`

| 功能 | 说明 |
|------|------|
| 生成邀请码 | 可设置使用次数上限、过期时间 |
| 查看列表 | 状态、已使用次数、关联注册用户 |
| 禁用/删除 | 停用或删除邀请码 |

#### 4.2.4 用量统计 `/agent/usage`

| 维度 | 说明 |
|------|------|
| 按用户 | 每个下游用户的消费明细 |
| 按时间 | 日 / 周 / 月汇总 |
| 按模型 | 各模型消费占比 |

---

## 5. API 接口设计

### 5.1 代理商接口 `/api/agent/*`

```
# 概览
GET  /api/agent/dashboard              # 获取概览数据

# 下游用户管理
GET  /api/agent/users                  # 用户列表
POST /api/agent/users                  # 创建用户
GET  /api/agent/users/:id              # 用户详情
PUT  /api/agent/users/:id              # 更新用户信息
POST /api/agent/users/:id/topup        # 充值
POST /api/agent/users/:id/deduct       # 扣款
PUT  /api/agent/users/:id/rate         # 设置倍率
PUT  /api/agent/users/:id/status       # 启用/禁用
POST /api/agent/users/:id/reset-password   # 重置密码

# 下游用户订阅
GET  /api/agent/users/:id/subscriptions        # 订阅列表
POST /api/agent/users/:id/subscriptions        # 分配订阅
PUT  /api/agent/users/:id/subscriptions/:sid   # 修改订阅

# 下游用户 API Key
GET    /api/agent/users/:id/api-keys           # API Key 列表
POST   /api/agent/users/:id/api-keys           # 创建 API Key
DELETE /api/agent/users/:id/api-keys/:kid      # 删除 API Key

# 邀请码
GET    /api/agent/invite-codes         # 邀请码列表
POST   /api/agent/invite-codes         # 生成邀请码
PUT    /api/agent/invite-codes/:id     # 更新邀请码
DELETE /api/agent/invite-codes/:id     # 删除邀请码

# 用量统计
GET  /api/agent/usage                  # 用量统计
```

### 5.2 权限控制

```
/api/admin/*   → 需要 role = admin
/api/agent/*   → 需要 role = agent
/api/user/*    → 需要 role = user 或 agent
/api/auth/*    → 公开接口
```

---

## 6. 管理员功能增强

### 6.1 用户管理页面增强

**新增筛选：**
- 按角色筛选（admin / agent / user）
- 按所属代理商筛选

**新增操作：**

| 操作 | 说明 |
|------|------|
| 设置为代理商 | 将 user 升级为 agent |
| 取消代理商 | 将 agent 降级为 user（需无下游用户） |
| 分配代理商 | 给用户指定所属代理商 |
| 解绑代理商 | 将下游用户改为平台直属 |

**列表新增显示：**
- 角色标签（代理商 badge）
- 所属代理商名称

### 6.2 新增代理商管理菜单

独立的「代理商管理」页面，展示：

| 内容 | 说明 |
|------|------|
| 代理商列表 | 所有 role=agent 的用户 |
| 基础信息 | 余额、下游用户数、可用分组 |
| 消费统计 | 下游用户总消费 |
| 快捷操作 | 充值、查看下游用户、设置可用分组 |

### 6.3 设置允许分组弹窗优化

**现状：** 只有"允许全部分组"选项

**改为：**
- 选项 1：允许全部分组（用户可使用任何非独占分组）
- 选项 2：指定分组（多选列表，列出所有 standard 类型分组）

---

## 7. 业务流程

### 7.1 用户注册流程

**场景 1：普通注册（无邀请码）**
```
用户注册 → role=user, agent_id=null → 平台直属用户
```

**场景 2：邀请码注册**
```
用户填写邀请码 → 校验有效性 → role=user, agent_id=代理商ID
              → 更新邀请码 used_count
```

**场景 3：代理商手动创建**
```
代理商创建用户 → role=user, agent_id=代理商ID
             → 可设置初始余额（从代理商余额划转）
```

### 7.2 计费流程

```
下游用户请求 API
    │
    ├─ 获取用户的 agent_rate
    │   └─ 若为 null，使用分组默认倍率
    │
    ├─ 计算费用 = 实际成本 × 倍率
    │
    └─ 扣减用户的 balance 或 subscription 额度
```

### 7.3 资金流转示例

```
1. 平台给代理商充值 $100
   代理商 balance: $0 → $100

2. 代理商给用户 A 充值 $20
   代理商 balance: $100 → $80
   用户 A balance: $0 → $20

3. 用户 A 消费，实际成本 $10，代理商设置倍率 1.5x
   用户 A 被扣: $10 × 1.5 = $15
   用户 A balance: $20 → $5

4. 代理商利润分析：
   - 代理商投入：$20（划转给用户）
   - 用户消费后剩余：$5（仍在用户账户）
   - 平台实际成本：$10
   - 代理商实际收益：$20 - $10 = $10
     （用户付出 $15 + 剩余 $5 = $20，平台成本 $10）
```

---

## 8. 前端页面设计

### 8.1 代理商后台风格

- 复用现有组件库和设计风格
- 独立路由 `/agent/*`
- 与用户后台、管理后台视觉一致

### 8.2 页面清单

| 页面 | 路由 | 说明 |
|------|------|------|
| 数据概览 | /agent | 首页 Dashboard |
| 用户列表 | /agent/users | 下游用户管理 |
| 用户详情 | /agent/users/:id | 用户详情与操作 |
| 邀请码管理 | /agent/invite-codes | 邀请码 CRUD |
| 用量统计 | /agent/usage | 消费报表 |

---

## 9. 实现优先级

### Phase 1：核心功能（必须）

1. 数据模型变更（users 表、agent_invite_codes 表）
2. 角色常量与权限中间件
3. 代理商后台基础页面（概览、用户管理）
4. 下游用户 CRUD 和充值功能
5. 管理员用户管理增强（角色筛选、分配代理商）

### Phase 2：完善功能

1. 邀请码管理
2. 注册流程支持邀请码
3. 代理商自定义倍率
4. 下游用户订阅管理
5. 设置允许分组弹窗优化

### Phase 3：运营功能

1. 代理商管理独立菜单
2. 用量统计页面
3. 数据导出功能

---

## 10. 风险与注意事项

| 风险点 | 应对措施 |
|--------|----------|
| 代理商设置低于成本的倍率 | 前端弹出警告提示，允许但提醒 |
| 代理商余额并发扣减 | 使用数据库事务 + 乐观锁 |
| 下游用户消费时代理商已无余额 | 不影响，下游用户用的是自己已充值的余额 |
| 代理商降级时有下游用户 | 校验禁止，需先转移或删除下游用户 |
| 用户分配给代理商时已有余额 | 余额保留，不做转移 |

---

## 11. 待讨论事项

- [ ] 代理商是否需要查看平台给自己的批发价/倍率？
- [ ] 是否需要代理商之间的数据隔离审计日志？
- [ ] 后续是否考虑对接支付，支持代理商在线充值？

---

## 附录：参考资料

- [One-API](https://github.com/songquanpeng/one-api) - LLM API 管理 & 分发系统
- [New-API](https://github.com/QuantumNous/new-api) - AI 模型聚合管理中转分发系统
- [CloseAI](https://doc.closeai-asia.com/tutorial/introduction.html) - 企业级商用 OpenAI 代理平台
