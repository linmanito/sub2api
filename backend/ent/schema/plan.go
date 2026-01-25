package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Plan holds the schema definition for the Plan entity (套餐).
type Plan struct {
	ent.Schema
}

func (Plan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "plans"},
	}
}

func (Plan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Plan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Comment("套餐名称"),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("套餐描述"),
		field.Int("price").
			Default(0).
			Comment("价格（分为单位）"),
		field.String("currency").
			MaxLen(10).
			Default("CNY").
			Comment("货币类型"),
		field.Int("validity_days").
			Default(30).
			Comment("有效天数"),
		field.Int("concurrency").
			Default(1).
			Comment("并发数限制"),
		field.JSON("features", []string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("特性描述列表"),
		field.String("icon").
			MaxLen(50).
			Default("standard").
			Comment("图标标识"),
		field.Bool("is_recommended").
			Default(false).
			Comment("是否推荐"),
		field.String("status").
			MaxLen(20).
			Default("active").
			Comment("状态: active/disabled"),
		field.Int("sort_order").
			Default(0).
			Comment("排序值（越小越靠前）"),
	}
}

func (Plan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", Group.Type).
			Comment("关联的分组"),
		edge.To("orders", Order.Type),
	}
}

func (Plan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("sort_order"),
		index.Fields("deleted_at"),
	}
}
