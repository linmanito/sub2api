package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Order holds the schema definition for the Order entity (订单).
type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "orders"},
	}
}

func (Order) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").
			Comment("用户ID"),
		field.Int64("plan_id").
			Comment("套餐ID"),
		field.String("status").
			MaxLen(20).
			Default("pending").
			Comment("状态: pending/confirmed/rejected/cancelled"),
		field.Int("amount").
			Default(0).
			Comment("订单金额（分为单位，快照）"),
		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("管理员备注"),
		field.Int64("confirmed_by").
			Optional().
			Nillable().
			Comment("确认人ID"),
		field.Time("confirmed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("确认时间"),
		field.Time("ordered_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("下单时间"),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("plan", Plan.Type).
			Ref("orders").
			Field("plan_id").
			Unique().
			Required(),
		edge.From("confirmed_by_user", User.Type).
			Ref("confirmed_orders").
			Field("confirmed_by").
			Unique(),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("plan_id"),
		index.Fields("status"),
		index.Fields("ordered_at"),
		index.Fields("confirmed_by"),
		index.Fields("deleted_at"),
	}
}
