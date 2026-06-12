package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/subhasundardas/gofar/framework/mixin"
)

// PartyMaster holds the schema definition for the PartyMaster entity.
type PartyMaster struct {
	ent.Schema
}

func (PartyMaster) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Base{},
	}
}

// Annotations of the User.
func (PartyMaster) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "party_master"},
	}
}

// Fields of the PartyMaster.
func (PartyMaster) Fields() []ent.Field {
	return []ent.Field{
		field.Int("ledger_id").Unique(),

		field.Enum("type").
			Values("CUSTOMER", "SUPPLIER", "BOTH"),

		field.String("display_name").
			NotEmpty(),

		field.String("legal_name").
			Optional().
			Nillable(),

		field.String("gst_no").
			Optional().
			Nillable(),

		field.String("pan_no").
			Optional().
			Nillable(),

		field.String("contact_person").
			Optional().
			Nillable(),

		field.String("mobile").
			Optional().
			Nillable(),

		field.String("phone").
			Optional().
			Nillable(),

		field.String("email").
			Optional().
			Nillable(),

		field.String("website").
			Optional().
			Nillable(),

		field.Float("credit_limit").
			Default(0),

		field.Int("credit_days").
			Default(0),

		field.Float("opening_balance").
			Default(0),

		field.String("address").
			Optional().
			Nillable(),

		field.String("city").
			Optional().
			Nillable(),

		field.String("state").
			Optional().
			Nillable(),

		field.String("country").
			Default("India"),

		field.String("pincode").
			Optional().
			Nillable(),
	}
}

func (PartyMaster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ledger_id").
			Unique(),

		index.Fields("type"),

		index.Fields("gst_no"),

		index.Fields("mobile"),

		index.Fields("email"),

		index.Fields("display_name"),
	}
}

// Edges of the PartyMaster.
func (PartyMaster) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ledger", Ledger.Type).
			Ref("party").
			Field("ledger_id").
			Required().
			Unique(),
	}
}
