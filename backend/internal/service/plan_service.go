package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// PlanService 套餐服务
type PlanService struct {
	planRepo  PlanRepository
	orderRepo OrderRepository
}

// NewPlanService 创建套餐服务
func NewPlanService(planRepo PlanRepository, orderRepo OrderRepository) *PlanService {
	return &PlanService{
		planRepo:  planRepo,
		orderRepo: orderRepo,
	}
}

// CreatePlanInput 创建套餐输入
type CreatePlanInput struct {
	Name          string
	Description   *string
	Price         int
	Currency      string
	ValidityDays  int
	Concurrency   int
	Features      []string
	Icon          string
	IsRecommended bool
	GroupIDs      []int64
	SortOrder     int
}

// UpdatePlanInput 更新套餐输入
type UpdatePlanInput struct {
	Name          *string
	Description   *string
	Price         *int
	Currency      *string
	ValidityDays  *int
	Concurrency   *int
	Features      []string
	Icon          *string
	IsRecommended *bool
	GroupIDs      []int64
	SortOrder     *int
}

// Create 创建套餐
func (s *PlanService) Create(ctx context.Context, input *CreatePlanInput) (*Plan, error) {
	plan := &Plan{
		Name:          input.Name,
		Description:   input.Description,
		Price:         input.Price,
		Currency:      input.Currency,
		ValidityDays:  input.ValidityDays,
		Concurrency:   input.Concurrency,
		Features:      input.Features,
		Icon:          input.Icon,
		IsRecommended: input.IsRecommended,
		Status:        StatusActive,
		SortOrder:     input.SortOrder,
	}

	if plan.Currency == "" {
		plan.Currency = "CNY"
	}
	if plan.Icon == "" {
		plan.Icon = "standard"
	}

	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}

	// 设置关联的分组
	if len(input.GroupIDs) > 0 {
		if err := s.planRepo.SetGroups(ctx, plan.ID, input.GroupIDs); err != nil {
			return nil, err
		}
	}

	// 重新获取完整数据
	return s.planRepo.GetByID(ctx, plan.ID)
}

// GetByID 根据 ID 获取套餐
func (s *PlanService) GetByID(ctx context.Context, id int64) (*Plan, error) {
	return s.planRepo.GetByID(ctx, id)
}

// Update 更新套餐
func (s *PlanService) Update(ctx context.Context, id int64, input *UpdatePlanInput) (*Plan, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		plan.Name = *input.Name
	}
	if input.Description != nil {
		plan.Description = input.Description
	}
	if input.Price != nil {
		plan.Price = *input.Price
	}
	if input.Currency != nil {
		plan.Currency = *input.Currency
	}
	if input.ValidityDays != nil {
		plan.ValidityDays = *input.ValidityDays
	}
	if input.Concurrency != nil {
		plan.Concurrency = *input.Concurrency
	}
	if input.Features != nil {
		plan.Features = input.Features
	}
	if input.Icon != nil {
		plan.Icon = *input.Icon
	}
	if input.IsRecommended != nil {
		plan.IsRecommended = *input.IsRecommended
	}
	if input.SortOrder != nil {
		plan.SortOrder = *input.SortOrder
	}

	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, err
	}

	// 更新关联的分组
	if input.GroupIDs != nil {
		if err := s.planRepo.SetGroups(ctx, id, input.GroupIDs); err != nil {
			return nil, err
		}
	}

	return s.planRepo.GetByID(ctx, id)
}

// Delete 删除套餐
func (s *PlanService) Delete(ctx context.Context, id int64) error {
	return s.planRepo.Delete(ctx, id)
}

// List 获取套餐列表（管理端）
func (s *PlanService) List(ctx context.Context, page, pageSize int, status string) ([]Plan, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.planRepo.List(ctx, params, status)
}

// ListActive 获取可用套餐列表（用户端）
func (s *PlanService) ListActive(ctx context.Context) ([]Plan, error) {
	return s.planRepo.ListActive(ctx)
}

// UpdateStatus 更新套餐状态
func (s *PlanService) UpdateStatus(ctx context.Context, id int64, status string) error {
	return s.planRepo.UpdateStatus(ctx, id, status)
}

// UpdateSortOrders 批量更新排序
func (s *PlanService) UpdateSortOrders(ctx context.Context, orders map[int64]int) error {
	return s.planRepo.UpdateSortOrders(ctx, orders)
}

// OrderService 订单服务
type OrderService struct {
	orderRepo           OrderRepository
	planRepo            PlanRepository
	subscriptionService *SubscriptionService
	userRepo            UserRepository
}

// NewOrderService 创建订单服务
func NewOrderService(
	orderRepo OrderRepository,
	planRepo PlanRepository,
	subscriptionService *SubscriptionService,
	userRepo UserRepository,
) *OrderService {
	return &OrderService{
		orderRepo:           orderRepo,
		planRepo:            planRepo,
		subscriptionService: subscriptionService,
		userRepo:            userRepo,
	}
}

// CreateOrderInput 创建订单输入
type CreateOrderInput struct {
	UserID int64
	PlanID int64
}

// Create 创建订单
func (s *OrderService) Create(ctx context.Context, input *CreateOrderInput) (*Order, error) {
	// 获取套餐信息
	plan, err := s.planRepo.GetByID(ctx, input.PlanID)
	if err != nil {
		return nil, err
	}

	// 检查套餐是否可用
	if !plan.IsActive() {
		return nil, ErrPlanNotFound
	}

	order := &Order{
		UserID:    input.UserID,
		PlanID:    input.PlanID,
		Status:    OrderStatusPending,
		Amount:    plan.Price,
		OrderedAt: time.Now(),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 重新获取完整数据
	return s.orderRepo.GetByID(ctx, order.ID)
}

// GetByID 根据 ID 获取订单
func (s *OrderService) GetByID(ctx context.Context, id int64) (*Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

// List 获取订单列表（管理端）
func (s *OrderService) List(ctx context.Context, page, pageSize int, userID *int64, status string) ([]Order, *pagination.PaginationResult, error) {
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.orderRepo.List(ctx, params, userID, status)
}

// ListByUserID 获取用户订单列表（用户端）
func (s *OrderService) ListByUserID(ctx context.Context, userID int64) ([]Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID)
}

// Cancel 取消订单（用户操作）
func (s *OrderService) Cancel(ctx context.Context, orderID int64, userID int64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 检查是否是该用户的订单
	if order.UserID != userID {
		return ErrOrderNotFound
	}

	// 检查订单状态
	if !order.CanCancel() {
		return ErrOrderPending
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, OrderStatusCancelled, nil)
}

// ConfirmInput 确认订单输入
type ConfirmInput struct {
	OrderID     int64
	ConfirmedBy int64
	Notes       *string
}

// Confirm 确认订单（管理员操作）
func (s *OrderService) Confirm(ctx context.Context, input *ConfirmInput) (*Order, error) {
	order, err := s.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}

	// 检查订单状态
	if !order.IsPending() {
		return nil, ErrOrderPending
	}

	// 获取套餐信息（包含关联的分组）
	plan, err := s.planRepo.GetByID(ctx, order.PlanID)
	if err != nil {
		return nil, err
	}

	// 为用户创建每个关联分组的订阅
	for _, group := range plan.Groups {
		_, _, err := s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
			UserID:       order.UserID,
			GroupID:      group.ID,
			ValidityDays: plan.ValidityDays,
			AssignedBy:   input.ConfirmedBy,
			Notes:        "订单确认自动创建",
		})
		if err != nil {
			return nil, err
		}
	}

	// 更新用户并发数
	if plan.Concurrency > 0 {
		if err := s.userRepo.UpdateConcurrency(ctx, order.UserID, plan.Concurrency); err != nil {
			return nil, err
		}
	}

	// 更新订单状态
	order.Status = OrderStatusConfirmed
	order.ConfirmedBy = &input.ConfirmedBy
	now := time.Now()
	order.ConfirmedAt = &now
	if input.Notes != nil {
		order.Notes = input.Notes
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return s.orderRepo.GetByID(ctx, input.OrderID)
}

// RejectInput 拒绝订单输入
type RejectInput struct {
	OrderID     int64
	ConfirmedBy int64
	Notes       *string
}

// Reject 拒绝订单（管理员操作）
func (s *OrderService) Reject(ctx context.Context, input *RejectInput) (*Order, error) {
	order, err := s.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}

	// 检查订单状态
	if !order.IsPending() {
		return nil, ErrOrderPending
	}

	// 更新订单状态
	order.Status = OrderStatusRejected
	order.ConfirmedBy = &input.ConfirmedBy
	now := time.Now()
	order.ConfirmedAt = &now
	if input.Notes != nil {
		order.Notes = input.Notes
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return s.orderRepo.GetByID(ctx, input.OrderID)
}
