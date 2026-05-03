package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SupplierProfile holds supplier onboarding and ownership metadata.
type SupplierProfile struct {
	ent.Schema
}

func (SupplierProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "supplier_profiles"},
	}
}

func (SupplierProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").
			Unique(),
		field.String("company_name").
			MaxLen(120).
			NotEmpty(),
		field.String("contact_name").
			MaxLen(100).
			Default(""),
		field.String("contact_email").
			MaxLen(255).
			Default(""),
		field.String("contact_phone").
			MaxLen(50).
			Default(""),
		field.String("status").
			MaxLen(20).
			Default(domain.SupplierStatusPending),
		field.JSON("settlement_config", map[string]any{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Default(func() map[string]any { return map[string]any{} }),
		field.String("notes").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.String("review_note").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.Time("reviewed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("reviewed_by").
			Optional().
			Nillable(),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupplierProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("supplier_profile").
			Field("user_id").
			Required().
			Unique(),
		edge.To("accounts", Account.Type),
		edge.To("usage_logs", UsageLog.Type),
	}
}

func (SupplierProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("user_id").Unique(),
	}
}
