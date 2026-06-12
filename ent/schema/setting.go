package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Setting holds the schema definition for the Setting entity.
type Setting struct {
	ent.Schema
}

// Fields of the Setting.
func (Setting) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty().Immutable(),
		field.String("value").Default(""),
		field.String("label").Default(""),        // human-readable name for UI
		field.String("group").Default("general"), // for grouping in UI
	}
}

func (Setting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key").Unique(),
	}
}

// Edges of the Setting.
func (Setting) Edges() []ent.Edge {
	return nil
}
