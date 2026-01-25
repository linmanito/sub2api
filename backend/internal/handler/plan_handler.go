package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PlanHandler 用户端套餐处理器
type PlanHandler struct {
	planService *service.PlanService
}

// NewPlanHandler 创建用户端套餐处理器
func NewPlanHandler(planService *service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

// PlanResponse 套餐响应
type PlanResponse struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	Description   *string       `json:"description"`
	Price         int           `json:"price"`
	PriceYuan     float64       `json:"price_yuan"`
	Currency      string        `json:"currency"`
	ValidityDays  int           `json:"validity_days"`
	Concurrency   int           `json:"concurrency"`
	Features      []string      `json:"features"`
	Icon          string        `json:"icon"`
	IsRecommended bool          `json:"is_recommended"`
	Groups        []GroupSimple `json:"groups"`
}

// GroupSimple 简化的分组信息
type GroupSimple struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Platform string  `json:"platform"`
}

// planToResponse 将服务层套餐转换为响应
func planToResponse(p *service.Plan) *PlanResponse {
	groups := make([]GroupSimple, 0, len(p.Groups))
	for _, g := range p.Groups {
		groups = append(groups, GroupSimple{
			ID:       g.ID,
			Name:     g.Name,
			Platform: g.Platform,
		})
	}

	return &PlanResponse{
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
		Groups:        groups,
	}
}

// List 获取可用套餐列表
// GET /api/v1/plans
func (h *PlanHandler) List(c *gin.Context) {
	plans, err := h.planService.ListActive(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result := make([]*PlanResponse, 0, len(plans))
	for i := range plans {
		result = append(result, planToResponse(&plans[i]))
	}

	response.Success(c, result)
}

// OrderHandler 用户端订单处理器
type OrderHandler struct {
	orderService   *service.OrderService
	settingService *service.SettingService
}

// NewOrderHandler 创建用户端订单处理器
func NewOrderHandler(orderService *service.OrderService, settingService *service.SettingService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		settingService: settingService,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	PlanID int64 `json:"plan_id" binding:"required"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	ID          int64        `json:"id"`
	PlanID      int64        `json:"plan_id"`
	Status      string       `json:"status"`
	Amount      int          `json:"amount"`
	AmountYuan  float64      `json:"amount_yuan"`
	OrderedAt   string       `json:"ordered_at"`
	ConfirmedAt *string      `json:"confirmed_at"`
	Plan        *PlanResponse `json:"plan,omitempty"`
}

// orderToResponse 将服务层订单转换为响应
func orderToResponse(o *service.Order) *OrderResponse {
	resp := &OrderResponse{
		ID:         o.ID,
		PlanID:     o.PlanID,
		Status:     o.Status,
		Amount:     o.Amount,
		AmountYuan: o.AmountYuan(),
		OrderedAt:  o.OrderedAt.Format("2006-01-02 15:04:05"),
	}

	if o.ConfirmedAt != nil {
		t := o.ConfirmedAt.Format("2006-01-02 15:04:05")
		resp.ConfirmedAt = &t
	}

	if o.Plan != nil {
		resp.Plan = planToResponse(o.Plan)
	}

	return resp
}

// Create 创建订单
// POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 获取当前用户 ID
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	order, err := h.orderService.Create(c.Request.Context(), &service.CreateOrderInput{
		UserID: subject.UserID,
		PlanID: req.PlanID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, orderToResponse(order))
}

// List 获取我的订单列表
// GET /api/v1/orders
func (h *OrderHandler) List(c *gin.Context) {
	// 获取当前用户 ID
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	orders, err := h.orderService.ListByUserID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result := make([]*OrderResponse, 0, len(orders))
	for i := range orders {
		result = append(result, orderToResponse(&orders[i]))
	}

	response.Success(c, result)
}

// Cancel 取消订单
// POST /api/v1/orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	// 获取当前用户 ID
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	if err := h.orderService.Cancel(c.Request.Context(), orderID, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Order cancelled successfully"})
}

// GetPaymentInfo 获取支付信息
// GET /api/v1/orders/:id/payment
func (h *OrderHandler) GetPaymentInfo(c *gin.Context) {
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	// 获取当前用户 ID
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	// 获取订单信息
	order, err := h.orderService.GetByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 检查是否是该用户的订单
	if order.UserID != subject.UserID {
		response.NotFound(c, "Order not found")
		return
	}

	// 获取支付配置
	allSettings, err := h.settingService.GetAllSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	paymentInfo := gin.H{
		"order":          orderToResponse(order),
		"alipay_qrcode":  allSettings.PaymentAlipayQrcode,
		"wechat_qrcode":  allSettings.PaymentWechatQrcode,
		"service_qrcode": allSettings.PaymentServiceQrcode,
		"payment_note":   allSettings.PaymentNote,
	}

	response.Success(c, paymentInfo)
}

// parseIDParam 解析 ID 参数
func parseIDParam(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}
