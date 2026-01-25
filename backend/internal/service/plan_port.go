package service

import (
	"context"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrPlanNotFound  = infraerrors.NotFound("PLAN_NOT_FOUND", "plan not found")
	ErrOrderNotFound = infraerrors.NotFound("ORDER_NOT_FOUND", "order not found")
	ErrOrderPending  = infraerrors.BadRequest("ORDER_NOT_PENDING", "order is not in pending status")
)

// PlanRepository 套餐仓储接口
type PlanRepository interface {
	Create(ctx context.Context, plan *Plan) error
	GetByID(ctx context.Context, id int64) (*Plan, error)
	Update(ctx context.Context, plan *Plan) error
	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, params pagination.PaginationParams, status string) ([]Plan, *pagination.PaginationResult, error)
	ListActive(ctx context.Context) ([]Plan, error)

	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateSortOrders(ctx context.Context, orders map[int64]int) error

	SetGroups(ctx context.Context, planID int64, groupIDs []int64) error
	GetGroups(ctx context.Context, planID int64) ([]Group, error)
}

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id int64) (*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, params pagination.PaginationParams, userID *int64, status string) ([]Order, *pagination.PaginationResult, error)
	ListByUserID(ctx context.Context, userID int64) ([]Order, error)

	UpdateStatus(ctx context.Context, id int64, status string, confirmedBy *int64) error
}
