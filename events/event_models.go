package events

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	SubjectID    primitive.ObjectID   `bson:"subject_id,omitempty"`
	ObjectID     primitive.ObjectID   `bson:"object_id,omitempty"`
	Type         string               `bson:"type,omitempty"`
	CreatedAt    time.Time            `bson:"created_at,omitempty"`
	Acknowledged []primitive.ObjectID `bson:"ackd"`
}
