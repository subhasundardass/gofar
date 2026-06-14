package schema

import (
	"github.com/subhasundardas/gofar/framework/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Journal_Line holds the schema definition for the Journal_Line entity.
type Journal_Line struct {
	ent.Schema
}

func (Journal_Line) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Base{},
	}
}

// Fields of the Journal_Line.
func (Journal_Line) Fields() []ent.Field {
	return []ent.Field{

		field.Int("journal_id"),

		field.Int("ledger_id"),

		field.Float("debit").
			Default(0),

		field.Float("credit").
			Default(0),

		field.String("description").
			Optional().
			Nillable(),

		field.String("reference_type").
			Optional().
			Nillable(),

		field.Int("reference_id").
			Optional().
			Nillable(),

		field.Int("line_no").
			Default(1),
	}
}

func (Journal_Line) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("journal_id"),

		index.Fields("ledger_id"),

		index.Fields("journal_id", "ledger_id"),
	}
}

// Edges of the Journal_Line.
func (Journal_Line) Edges() []ent.Edge {
	return []ent.Edge{

		edge.From("journal", Journal.Type).
			Ref("lines").
			Field("journal_id").
			Required().
			Unique(),

		edge.From("ledger", Ledger.Type).
			Ref("journal_lines"). // make sure Ledger has edge.To("journal_lines", Journal_Line.Type)
			Field("ledger_id").
			Required().
			Unique(),
	}
}
