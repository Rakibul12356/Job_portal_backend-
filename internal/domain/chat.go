package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatRoom struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JobID      primitive.ObjectID `bson:"jobId" json:"jobId"`
	SeekerID   primitive.ObjectID `bson:"seekerId" json:"seekerId"`
	EmployerID primitive.ObjectID `bson:"employerId" json:"employerId"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type ChatMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    primitive.ObjectID `bson:"roomId" json:"roomId"`
	SenderID  primitive.ObjectID `bson:"senderId" json:"senderId"`
	Message   string             `bson:"message" json:"message"`
	IsRead    bool               `bson:"isRead" json:"isRead"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
