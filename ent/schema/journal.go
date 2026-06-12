package schema

import (
	"github.com/subhasundardas/gofar/framework/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Journal holds the schema definition for the Journal entity.
type Journal struct {
	ent.Schema
}

func (Journal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Base{},
	}
}

// Fields of the Journal.
func (Journal) Fields() []ent.Field {
	return []ent.Field{

		field.Time("date"),

		field.String("voucher_type").
			NotEmpty(), // JV, PV, RV, CV, SV, PR

		field.String("voucher_no").
			NotEmpty(),

		field.Time("voucher_date"),

		field.String("reference_no").
			Optional().
			Nillable(),

		field.String("external_ref").
			Optional().
			Nillable(),

		field.Enum("journal_status").
			Values(
				"DRAFT",
				"APPROVED",
				"POSTED",
				"CANCELLED",
			).
			Default("DRAFT"), // DRAFT, POSTED, CANCELLED

		field.Int("approved_by").
			Optional().
			Nillable(),

		field.Int("financial_year_id").
			Optional().
			Nillable(),

		field.String("narration").
			Optional().
			Nillable(),

		field.String("source_module").
			Optional().
			Nillable(), // sales,purchase,payroll

		field.String("source_type").
			Optional().
			Nillable(),

		field.Int("source_id").
			Optional().
			Nillable(),

		field.Float("total_debit").
			Default(0),

		field.Float("total_credit").
			Default(0),
	}
}

func (Journal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("date"),

		index.Fields("voucher_type"),

		index.Fields("voucher_no").
			Unique(),

		index.Fields("status"),

		index.Fields("date", "voucher_type"),
	}
}

// Edges of the Journal.
func (Journal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("lines", Journal_Line.Type),
	}
}
