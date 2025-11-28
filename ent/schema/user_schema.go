package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("google_drive_id").NotEmpty(),
		field.String("email").NotEmpty().Unique(),
		field.String("name").NotEmpty().Unique(),
		field.String("access_token").NotEmpty(),
		field.String("token_type").NotEmpty(),
		field.String("refresh_token").NotEmpty().Unique(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("files", MediaFile.Type),
	}
}
