package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/order"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type orderRepository struct {
	client *dbent.Client
}

func NewOrderRepository(client *dbent.Client) service.OrderRepository {
	return &orderRepository{client: client}
}

func (r *orderRepository) Create(ctx context.Context, orderIn *service.Order) error {
	client := clientFromContext(ctx, r.client)

	builder := client.Order.Create().
		SetUserID(orderIn.UserID).
		SetPlanID(orderIn.PlanID).
		SetStatus(orderIn.Status).
		SetAmount(orderIn.Amount)

	if orderIn.Notes != nil {
		builder.SetNotes(*orderIn.Notes)
	}
	if !orderIn.OrderedAt.IsZero() {
		builder.SetOrderedAt(orderIn.OrderedAt)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}

	orderIn.ID = created.ID
	orderIn.OrderedAt = created.OrderedAt
	orderIn.CreatedAt = created.CreatedAt
	orderIn.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*service.Order, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Order.Query().
		Where(order.IDEQ(id)).
		WithUser().
		WithPlan(func(q *dbent.PlanQuery) {
			q.WithGroups()
		}).
		WithConfirmedByUser().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOrderNotFound, nil)
	}
	return orderEntityToService(m), nil
}

func (r *orderRepository) Update(ctx context.Context, orderIn *service.Order) error {
	client := clientFromContext(ctx, r.client)

	builder := client.Order.UpdateOneID(orderIn.ID).
		SetStatus(orderIn.Status).
		SetAmount(orderIn.Amount)

	if orderIn.Notes != nil {
		builder.SetNotes(*orderIn.Notes)
	} else {
		builder.ClearNotes()
	}
	if orderIn.ConfirmedBy != nil {
		builder.SetConfirmedBy(*orderIn.ConfirmedBy)
	}
	if orderIn.ConfirmedAt != nil {
		builder.SetConfirmedAt(*orderIn.ConfirmedAt)
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrOrderNotFound, nil)
	}

	orderIn.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *orderRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.Order.DeleteOneID(id).Exec(ctx)
}

func (r *orderRepository) List(ctx context.Context, params pagination.PaginationParams, userID *int64, status string) ([]service.Order, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)

	query := client.Order.Query().
		WithUser().
		WithPlan().
		WithConfirmedByUser()

	if userID != nil {
		query = query.Where(order.UserIDEQ(*userID))
	}
	if status != "" {
		query = query.Where(order.StatusEQ(status))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := query.
		Order(dbent.Desc(order.FieldOrderedAt)).
		Offset(params.Offset()).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	result := make([]service.Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, *orderEntityToService(o))
	}

	paginationResult := paginationResultFromTotal(int64(total), params)
	return result, paginationResult, nil
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID int64) ([]service.Order, error) {
	client := clientFromContext(ctx, r.client)

	orders, err := client.Order.Query().
		Where(order.UserIDEQ(userID)).
		WithPlan().
		Order(dbent.Desc(order.FieldOrderedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]service.Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, *orderEntityToService(o))
	}
	return result, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status string, confirmedBy *int64) error {
	client := clientFromContext(ctx, r.client)

	builder := client.Order.UpdateOneID(id).SetStatus(status)

	if confirmedBy != nil {
		builder.SetConfirmedBy(*confirmedBy)
		now := time.Now()
		builder.SetConfirmedAt(now)
	}

	return builder.Exec(ctx)
}

// orderEntityToService 将 ent Order 实体转换为 service Order
func orderEntityToService(m *dbent.Order) *service.Order {
	if m == nil {
		return nil
	}

	o := &service.Order{
		ID:          m.ID,
		UserID:      m.UserID,
		PlanID:      m.PlanID,
		Status:      m.Status,
		Amount:      m.Amount,
		Notes:       m.Notes,
		ConfirmedBy: m.ConfirmedBy,
		ConfirmedAt: m.ConfirmedAt,
		OrderedAt:   m.OrderedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	// 加载关联的用户
	if m.Edges.User != nil {
		o.User = userEntityToService(m.Edges.User)
	}

	// 加载关联的套餐
	if m.Edges.Plan != nil {
		o.Plan = planEntityToService(m.Edges.Plan)
	}

	// 加载确认人
	if m.Edges.ConfirmedByUser != nil {
		o.ConfirmedByUser = userEntityToService(m.Edges.ConfirmedByUser)
	}

	return o
}
