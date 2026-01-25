-- 044_add_plans_and_orders.sql
-- 添加套餐和订单功能

-- 创建 plans 表（套餐）
CREATE TABLE IF NOT EXISTS plans (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    validity_days INTEGER NOT NULL DEFAULT 30,
    concurrency INTEGER NOT NULL DEFAULT 1,
    features JSONB DEFAULT '[]'::jsonb,
    icon VARCHAR(50) NOT NULL DEFAULT 'standard',
    is_recommended BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- plans 表索引
CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status);
CREATE INDEX IF NOT EXISTS idx_plans_sort_order ON plans(sort_order);
CREATE INDEX IF NOT EXISTS idx_plans_deleted_at ON plans(deleted_at);

-- 创建 plan_groups 表（套餐-分组关联，多对多）
CREATE TABLE IF NOT EXISTS plan_groups (
    plan_id BIGINT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (plan_id, group_id)
);

-- plan_groups 表索引
CREATE INDEX IF NOT EXISTS idx_plan_groups_plan_id ON plan_groups(plan_id);
CREATE INDEX IF NOT EXISTS idx_plan_groups_group_id ON plan_groups(group_id);

-- 创建 orders 表（订单）
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id BIGINT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    confirmed_by BIGINT REFERENCES users(id),
    confirmed_at TIMESTAMPTZ,
    ordered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- orders 表索引
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_plan_id ON orders(plan_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_ordered_at ON orders(ordered_at);
CREATE INDEX IF NOT EXISTS idx_orders_confirmed_by ON orders(confirmed_by);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);

-- 添加注释
COMMENT ON TABLE plans IS '套餐表';
COMMENT ON COLUMN plans.name IS '套餐名称';
COMMENT ON COLUMN plans.description IS '套餐描述';
COMMENT ON COLUMN plans.price IS '价格（分为单位）';
COMMENT ON COLUMN plans.currency IS '货币类型';
COMMENT ON COLUMN plans.validity_days IS '有效天数';
COMMENT ON COLUMN plans.concurrency IS '并发数限制';
COMMENT ON COLUMN plans.features IS '特性描述列表';
COMMENT ON COLUMN plans.icon IS '图标标识';
COMMENT ON COLUMN plans.is_recommended IS '是否推荐';
COMMENT ON COLUMN plans.status IS '状态: active/disabled';
COMMENT ON COLUMN plans.sort_order IS '排序值（越小越靠前）';

COMMENT ON TABLE plan_groups IS '套餐-分组关联表';

COMMENT ON TABLE orders IS '订单表';
COMMENT ON COLUMN orders.user_id IS '用户ID';
COMMENT ON COLUMN orders.plan_id IS '套餐ID';
COMMENT ON COLUMN orders.status IS '状态: pending/confirmed/rejected/cancelled';
COMMENT ON COLUMN orders.amount IS '订单金额（分为单位，快照）';
COMMENT ON COLUMN orders.notes IS '管理员备注';
COMMENT ON COLUMN orders.confirmed_by IS '确认人ID';
COMMENT ON COLUMN orders.confirmed_at IS '确认时间';
COMMENT ON COLUMN orders.ordered_at IS '下单时间';
