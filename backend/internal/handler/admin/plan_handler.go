package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PlanHandler 管理端套餐处理器
type PlanHandler struct {
	planService *service.PlanService
}

// NewPlanHandler 创建管理端套餐处理器
func NewPlanHandler(planService *service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

// CreatePlanRequest 创建套餐请求
type CreatePlanRequest struct {
	Name          string   `json:"name" binding:"required,max=100"`
	Description   *string  `json:"description"`
	Price         int      `json:"price" binding:"min=0"`
	Currency      string   `json:"currency"`
	ValidityDays  int      `json:"validity_days" binding:"min=1"`
	Concurrency   int      `json:"concurrency" binding:"min=1"`
	Features      []string `json:"features"`
	Icon          string   `json:"icon"`
	IsRecommended bool     `json:"is_recommended"`
	GroupIDs      []int64  `json:"group_ids"`
	SortOrder     int      `json:"sort_order"`
}

// UpdatePlanRequest 更新套餐请求
type UpdatePlanRequest struct {
	Name          *string  `json:"name" binding:"omitempty,max=100"`
	Description   *string  `json:"description"`
	Price         *int     `json:"price" binding:"omitempty,min=0"`
	Currency      *string  `json:"currency"`
	ValidityDays  *int     `json:"validity_days" binding:"omitempty,min=1"`
	Concurrency   *int     `json:"concurrency" binding:"omitempty,min=1"`
	Features      []string `json:"features"`
	Icon          *string  `json:"icon"`
	IsRecommended *bool    `json:"is_recommended"`
	GroupIDs      []int64  `json:"group_ids"`
	SortOrder     *int     `json:"sort_order"`
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// UpdateSortOrdersRequest 更新排序请求
type UpdateSortOrdersRequest struct {
	Orders map[int64]int `json:"orders" binding:"required"`
}

// AdminPlanResponse 管理端套餐响应
type AdminPlanResponse struct {
	ID            int64                `json:"id"`
	Name          string               `json:"name"`
	Description   *string              `json:"description"`
	Price         int                  `json:"price"`
	PriceYuan     float64              `json:"price_yuan"`
	Currency      string               `json:"currency"`
	ValidityDays  int                  `json:"validity_days"`
	Concurrency   int                  `json:"concurrency"`
	Features      []string             `json:"features"`
	Icon          string               `json:"icon"`
	IsRecommended bool                 `json:"is_recommended"`
	Status        string               `json:"status"`
	SortOrder     int                  `json:"sort_order"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
	Groups        []AdminGroupSimple   `json:"groups"`
}

// AdminGroupSimple 简化的分组信息
type AdminGroupSimple struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

// planToAdminResponse 将服务层套餐转换为管理端响应
func planToAdminResponse(p *service.Plan) *AdminPlanResponse {
	groups := make([]AdminGroupSimple, 0, len(p.Groups))
	for _, g := range p.Groups {
		groups = append(groups, AdminGroupSimple{
			ID:       g.ID,
			Name:     g.Name,
			Platform: g.Platform,
		})
	}

	return &AdminPlanResponse{
		ID:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price,
		PriceYuan:     p.PriceYuan(),
		Currency:      p.Currency,
		ValidityDays:  p.ValidityDays,
		Concurrency:   p.Concurrency,
		Features:      p.Features,
		Icon:          p.Icon,
		IsRecommended: p.IsRecommended,
		Status:        p.Status,
		SortOrder:     p.SortOrder,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02 15:04:05"),
		Groups:        groups,
	}
}

// List 获取套餐列表
// GET /api/v1/admin/plans
func (h *PlanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	plans, pagination, err := h.planService.List(c.Request.Context(), page, pageSize, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result := make([]*AdminPlanResponse, 0, len(plans))
	for i := range plans {
		result = append(result, planToAdminResponse(&plans[i]))
	}

	response.PaginatedWithResult(c, result, toResponsePagination(pagination))
}

// GetByID 获取套餐详情
// GET /api/v1/admin/plans/:id
func (h *PlanHandler) GetByID(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.planService.GetByID(c.Request.Context(), planID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, planToAdminResponse(plan))
}

// Create 创建套餐
// POST /api/v1/admin/plans
func (h *PlanHandler) Create(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.Create(c.Request.Context(), &service.CreatePlanInput{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Currency:      req.Currency,
		ValidityDays:  req.ValidityDays,
		Concurrency:   req.Concurrency,
		Features:      req.Features,
		Icon:          req.Icon,
		IsRecommended: req.IsRecommended,
		GroupIDs:      req.GroupIDs,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, planToAdminResponse(plan))
}

// Update 更新套餐
// PUT /api/v1/admin/plans/:id
func (h *PlanHandler) Update(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.Update(c.Request.Context(), planID, &service.UpdatePlanInput{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Currency:      req.Currency,
		ValidityDays:  req.ValidityDays,
		Concurrency:   req.Concurrency,
		Features:      req.Features,
		Icon:          req.Icon,
		IsRecommended: req.IsRecommended,
		GroupIDs:      req.GroupIDs,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, planToAdminResponse(plan))
}

// Delete 删除套餐
// DELETE /api/v1/admin/plans/:id
func (h *PlanHandler) Delete(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.planService.Delete(c.Request.Context(), planID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Plan deleted successfully"})
}

// UpdateStatus 更新套餐状态
// PUT /api/v1/admin/plans/:id/status
func (h *PlanHandler) UpdateStatus(c *gin.Context) {
	planID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.planService.UpdateStatus(c.Request.Context(), planID, req.Status); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Plan status updated successfully"})
}

// UpdateSortOrders 批量更新排序
// PUT /api/v1/admin/plans/sort
func (h *PlanHandler) UpdateSortOrders(c *gin.Context) {
	var req UpdateSortOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.planService.UpdateSortOrders(c.Request.Context(), req.Orders); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Sort orders updated successfully"})
}

// OrderHandler 管理端订单处理器
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler 创建管理端订单处理器
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// ConfirmOrderRequest 确认订单请求
type ConfirmOrderRequest struct {
	Notes *string `json:"notes"`
}

// RejectOrderRequest 拒绝订单请求
type RejectOrderRequest struct {
	Notes *string `json:"notes"`
}

// AdminOrderResponse 管理端订单响应
type AdminOrderResponse struct {
	ID              int64              `json:"id"`
	UserID          int64              `json:"user_id"`
	PlanID          int64              `json:"plan_id"`
	Status          string             `json:"status"`
	Amount          int                `json:"amount"`
	AmountYuan      float64            `json:"amount_yuan"`
	Notes           *string            `json:"notes"`
	ConfirmedBy     *int64             `json:"confirmed_by"`
	ConfirmedAt     *string            `json:"confirmed_at"`
	OrderedAt       string             `json:"ordered_at"`
	CreatedAt       string             `json:"created_at"`
	User            *AdminUserSimple   `json:"user,omitempty"`
	Plan            *AdminPlanResponse `json:"plan,omitempty"`
	ConfirmedByUser *AdminUserSimple   `json:"confirmed_by_user,omitempty"`
}

// AdminUserSimple 简化的用户信息
type AdminUserSimple struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// orderToAdminResponse 将服务层订单转换为管理端响应
func orderToAdminResponse(o *service.Order) *AdminOrderResponse {
	resp := &AdminOrderResponse{
		ID:          o.ID,
		UserID:      o.UserID,
		PlanID:      o.PlanID,
		Status:      o.Status,
		Amount:      o.Amount,
		AmountYuan:  o.AmountYuan(),
		Notes:       o.Notes,
		ConfirmedBy: o.ConfirmedBy,
		OrderedAt:   o.OrderedAt.Format("2006-01-02 15:04:05"),
		CreatedAt:   o.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if o.ConfirmedAt != nil {
		t := o.ConfirmedAt.Format("2006-01-02 15:04:05")
		resp.ConfirmedAt = &t
	}

	if o.User != nil {
		resp.User = &AdminUserSimple{
			ID:       o.User.ID,
			Email:    o.User.Email,
			Username: o.User.Username,
		}
	}

	if o.Plan != nil {
		resp.Plan = planToAdminResponse(o.Plan)
	}

	if o.ConfirmedByUser != nil {
		resp.ConfirmedByUser = &AdminUserSimple{
			ID:       o.ConfirmedByUser.ID,
			Email:    o.ConfirmedByUser.Email,
			Username: o.ConfirmedByUser.Username,
		}
	}

	return resp
}

// List 获取订单列表
// GET /api/v1/admin/orders
func (h *OrderHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = &id
		}
	}

	orders, pagination, err := h.orderService.List(c.Request.Context(), page, pageSize, userID, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result := make([]*AdminOrderResponse, 0, len(orders))
	for i := range orders {
		result = append(result, orderToAdminResponse(&orders[i]))
	}

	response.PaginatedWithResult(c, result, toResponsePagination(pagination))
}

// GetByID 获取订单详情
// GET /api/v1/admin/orders/:id
func (h *OrderHandler) GetByID(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, orderToAdminResponse(order))
}

// Confirm 确认订单
// POST /api/v1/admin/orders/:id/confirm
func (h *OrderHandler) Confirm(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req ConfirmOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// notes 是可选的，忽略绑定错误
		req = ConfirmOrderRequest{}
	}

	// 获取当前管理员 ID
	adminID := getAdminIDFromContext(c)

	order, err := h.orderService.Confirm(c.Request.Context(), &service.ConfirmInput{
		OrderID:     orderID,
		ConfirmedBy: adminID,
		Notes:       req.Notes,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, orderToAdminResponse(order))
}

// Reject 拒绝订单
// POST /api/v1/admin/orders/:id/reject
func (h *OrderHandler) Reject(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req RejectOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// notes 是可选的，忽略绑定错误
		req = RejectOrderRequest{}
	}

	// 获取当前管理员 ID
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	adminID := int64(0)
	if ok {
		adminID = subject.UserID
	}

	order, err := h.orderService.Reject(c.Request.Context(), &service.RejectInput{
		OrderID:     orderID,
		ConfirmedBy: adminID,
		Notes:       req.Notes,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, orderToAdminResponse(order))
}
