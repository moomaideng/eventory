package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DummyResource represents an example document stored in MongoDB.
type DummyResource struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string        `bson:"title" json:"title"`
	Content   string        `bson:"content" json:"content"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updatedAt"`
}
