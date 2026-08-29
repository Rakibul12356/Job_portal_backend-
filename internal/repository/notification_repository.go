package repository

import (
	"context"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif *domain.Notification) error
	FindAllByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, notifID primitive.ObjectID) error
}

type mongoNotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) NotificationRepository {
	return &mongoNotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *mongoNotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	notif.CreatedAt = time.Now()
	if notif.ID.IsZero() {
		notif.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, notif)
	return err
}

func (r *mongoNotificationRepository) FindAllByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Notification, error) {
	notifs := make([]*domain.Notification, 0)
	opts := options.Find().SetSort(bson.M{"createdAt": -1}) // Return newest notifications first

	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var notif domain.Notification
		if err := cursor.Decode(&notif); err != nil {
			return nil, err
		}
		notifs = append(notifs, &notif)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return notifs, nil
}

func (r *mongoNotificationRepository) MarkAsRead(ctx context.Context, notifID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": notifID},
		bson.M{"$set": bson.M{"isRead": true}},
	)
	return err
}
