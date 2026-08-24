package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedJob struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	JobID     primitive.ObjectID `bson:"jobId" json:"jobId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
