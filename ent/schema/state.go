package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// State holds the schema definition for the State entity.
type State struct {
	ent.Schema
}

// Fields of the State.
func (State) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100),

		field.String("code").
			MaxLen(10).
			Optional(),

		field.String("gst_code").
			MaxLen(5).
			Optional(),

		field.Int("country_id"),
	}
}

// Edges of the State.
func (State) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("country", Country.Type).
			Ref("states").
			Field("country_id").
			Unique().
			Required(),
	}
}

func (State) Indexes() []ent.Index {
	return []ent.Index{
		// Equivalent to GORM: index on country_id
		index.Fields("country_id"),

		// Prevent duplicate state names within same country
		index.Fields("country_id", "name").
			Unique(),

		// Optional: quick lookup by GST code
		index.Fields("gst_code"),
	}
}
