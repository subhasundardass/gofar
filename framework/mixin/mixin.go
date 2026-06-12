package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Base is the standard mixin for all ent schemas.
// It provides id, created_at, and updated_at fields.
//
// Usage in any schema:
//
//	func (MyEntity) Mixin() []ent.Mixin {
//	    return []ent.Mixin{
//	        coremixin.Base{},
//	    }
//	}
type Base struct {
	mixin.Schema
}

func (Base) Fields() []ent.Field {
	return []ent.Field{
		field.Int8("status").
			Default(1).
			Comment("Data Status (0:created, 1:active, 98:deleted)"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Timestamp when the record was created."),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Timestamp when the record was last updated."),
	}
}
