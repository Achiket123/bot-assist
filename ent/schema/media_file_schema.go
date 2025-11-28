package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type MediaFile struct {
	ent.Schema
}

func (MediaFile) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("url").Unique(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (MediaFile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user_id", User.Type).Ref("files"),
	}
}
