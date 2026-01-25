package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/plan"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type planRepository struct {
	client *dbent.Client
}

func NewPlanRepository(client *dbent.Client) service.PlanRepository {
	return &planRepository{client: client}
}

func (r *planRepository) Create(ctx context.Context, planIn *service.Plan) error {
	client := clientFromContext(ctx, r.client)

	builder := client.Plan.Create().
		SetName(planIn.Name).
		SetPrice(planIn.Price).
		SetCurrency(planIn.Currency).
		SetValidityDays(planIn.ValidityDays).
		SetConcurrency(planIn.Concurrency).
		SetIcon(planIn.Icon).
		SetIsRecommended(planIn.IsRecommended).
		SetStatus(planIn.Status).
		SetSortOrder(planIn.SortOrder)

	if planIn.Description != nil {
		builder.SetDescription(*planIn.Description)
	}
	if planIn.Features != nil {
		builder.SetFeatures(planIn.Features)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}

	planIn.ID = created.ID
	planIn.CreatedAt = created.CreatedAt
	planIn.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *planRepository) GetByID(ctx context.Context, id int64) (*service.Plan, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.Plan.Query().
		Where(plan.IDEQ(id)).
		WithGroups().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPlanNotFound, nil)
	}
	return planEntityToService(m), nil
}

func (r *planRepository) Update(ctx context.Context, planIn *service.Plan) error {
	client := clientFromContext(ctx, r.client)

	builder := client.Plan.UpdateOneID(planIn.ID).
		SetName(planIn.Name).
		SetPrice(planIn.Price).
		SetCurrency(planIn.Currency).
		SetValidityDays(planIn.ValidityDays).
		SetConcurrency(planIn.Concurrency).
		SetIcon(planIn.Icon).
		SetIsRecommended(planIn.IsRecommended).
		SetStatus(planIn.Status).
		SetSortOrder(planIn.SortOrder)

	if planIn.Description != nil {
		builder.SetDescription(*planIn.Description)
	} else {
		builder.ClearDescription()
	}
	if planIn.Features != nil {
		builder.SetFeatures(planIn.Features)
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrPlanNotFound, nil)
	}

	planIn.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *planRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.Plan.DeleteOneID(id).Exec(ctx)
}

func (r *planRepository) List(ctx context.Context, params pagination.PaginationParams, status string) ([]service.Plan, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)

	query := client.Plan.Query().WithGroups()

	if status != "" {
		query = query.Where(plan.StatusEQ(status))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	plans, err := query.
		Order(dbent.Asc(plan.FieldSortOrder), dbent.Desc(plan.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	result := make([]service.Plan, 0, len(plans))
	for _, p := range plans {
		result = append(result, *planEntityToService(p))
	}

	paginationResult := paginationResultFromTotal(int64(total), params)
	return result, paginationResult, nil
}

func (r *planRepository) ListActive(ctx context.Context) ([]service.Plan, error) {
	client := clientFromContext(ctx, r.client)

	plans, err := client.Plan.Query().
		Where(plan.StatusEQ(service.StatusActive)).
		WithGroups().
		Order(dbent.Asc(plan.FieldSortOrder), dbent.Desc(plan.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]service.Plan, 0, len(plans))
	for _, p := range plans {
		result = append(result, *planEntityToService(p))
	}
	return result, nil
}

func (r *planRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	client := clientFromContext(ctx, r.client)
	return client.Plan.UpdateOneID(id).SetStatus(status).Exec(ctx)
}

func (r *planRepository) UpdateSortOrders(ctx context.Context, orders map[int64]int) error {
	client := clientFromContext(ctx, r.client)

	// 使用事务批量更新
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	for planID, sortOrder := range orders {
		if err := tx.Plan.UpdateOneID(planID).SetSortOrder(sortOrder).Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *planRepository) SetGroups(ctx context.Context, planID int64, groupIDs []int64) error {
	client := clientFromContext(ctx, r.client)

	// 先清空现有关联，再添加新关联
	err := client.Plan.UpdateOneID(planID).ClearGroups().Exec(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrPlanNotFound, nil)
	}

	if len(groupIDs) == 0 {
		return nil
	}

	// 添加新关联
	return client.Plan.UpdateOneID(planID).AddGroupIDs(groupIDs...).Exec(ctx)
}

func (r *planRepository) GetGroups(ctx context.Context, planID int64) ([]service.Group, error) {
	client := clientFromContext(ctx, r.client)

	groups, err := client.Group.Query().
		Where(group.HasPlansWith(plan.IDEQ(planID))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]service.Group, 0, len(groups))
	for _, g := range groups {
		result = append(result, *groupEntityToService(g))
	}
	return result, nil
}

// planEntityToService 将 ent Plan 实体转换为 service Plan
func planEntityToService(m *dbent.Plan) *service.Plan {
	if m == nil {
		return nil
	}

	p := &service.Plan{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Price:         m.Price,
		Currency:      m.Currency,
		ValidityDays:  m.ValidityDays,
		Concurrency:   m.Concurrency,
		Features:      m.Features,
		Icon:          m.Icon,
		IsRecommended: m.IsRecommended,
		Status:        m.Status,
		SortOrder:     m.SortOrder,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}

	// 加载关联的分组
	if m.Edges.Groups != nil {
		p.Groups = make([]service.Group, 0, len(m.Edges.Groups))
		for _, g := range m.Edges.Groups {
			p.Groups = append(p.Groups, *groupEntityToService(g))
		}
	}

	return p
}
