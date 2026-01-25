package service

import "time"

// Plan 套餐模型
type Plan struct {
	ID            int64
	Name          string
	Description   *string
	Price         int // 价格（分为单位）
	Currency      string
	ValidityDays  int
	Concurrency   int
	Features      []string
	Icon          string
	IsRecommended bool
	Status        string // active/disabled
	SortOrder     int

	CreatedAt time.Time
	UpdatedAt time.Time

	Groups []Group // 关联的分组
}

// IsActive 检查套餐是否启用
func (p *Plan) IsActive() bool {
	return p.Status == StatusActive
}

// PriceYuan 获取价格（元）
func (p *Plan) PriceYuan() float64 {
	return float64(p.Price) / 100
}

// Order 订单模型
type Order struct {
	ID          int64
	UserID      int64
	PlanID      int64
	Status      string // pending/confirmed/rejected/cancelled
	Amount      int    // 订单金额（分为单位，快照）
	Notes       *string
	ConfirmedBy *int64
	ConfirmedAt *time.Time
	OrderedAt   time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	User            *User
	Plan            *Plan
	ConfirmedByUser *User
}

// 订单状态常量
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusRejected  = "rejected"
	OrderStatusCancelled = "cancelled"
)

// IsPending 检查订单是否待审核
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// CanCancel 检查订单是否可取消
func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPending
}

// AmountYuan 获取金额（元）
func (o *Order) AmountYuan() float64 {
	return float64(o.Amount) / 100
}
