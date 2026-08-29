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

type ChatRepository interface {
	CreateRoom(ctx context.Context, room *domain.ChatRoom) error
	FindRoomByID(ctx context.Context, id primitive.ObjectID) (*domain.ChatRoom, error)
	FindRoomByParticipants(ctx context.Context, jobID, seekerID, employerID primitive.ObjectID) (*domain.ChatRoom, error)
	ListRoomsForUser(ctx context.Context, userID primitive.ObjectID, role string) ([]domain.ChatRoom, error)
	SaveMessage(ctx context.Context, msg *domain.ChatMessage) error
	GetMessagesByRoom(ctx context.Context, roomID primitive.ObjectID, limit, offset int64) ([]domain.ChatMessage, error)
	MarkMessagesAsRead(ctx context.Context, roomID, readerID primitive.ObjectID) error
	MarkAllMessagesAsDelivered(ctx context.Context, recipientID primitive.ObjectID) ([]primitive.ObjectID, error)
	GetLastMessage(ctx context.Context, roomID primitive.ObjectID) (*domain.ChatMessage, error)
	GetUnreadCount(ctx context.Context, roomID, userID primitive.ObjectID) (int64, error)
}

type mongoChatRepository struct {
	roomsCol    *mongo.Collection
	messagesCol *mongo.Collection
}

func NewChatRepository(db *mongo.Database) ChatRepository {
	return &mongoChatRepository{
		roomsCol:    db.Collection("chat_rooms"),
		messagesCol: db.Collection("chat_messages"),
	}
}

func (r *mongoChatRepository) CreateRoom(ctx context.Context, room *domain.ChatRoom) error {
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	if room.ID.IsZero() {
		room.ID = primitive.NewObjectID()
	}
	_, err := r.roomsCol.InsertOne(ctx, room)
	return err
}

func (r *mongoChatRepository) FindRoomByID(ctx context.Context, id primitive.ObjectID) (*domain.ChatRoom, error) {
	var room domain.ChatRoom
	err := r.roomsCol.FindOne(ctx, bson.M{"_id": id}).Decode(&room)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *mongoChatRepository) FindRoomByParticipants(ctx context.Context, jobID, seekerID, employerID primitive.ObjectID) (*domain.ChatRoom, error) {
	var room domain.ChatRoom
	err := r.roomsCol.FindOne(ctx, bson.M{
		"jobId":      jobID,
		"seekerId":   seekerID,
		"employerId": employerID,
	}).Decode(&room)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *mongoChatRepository) ListRoomsForUser(ctx context.Context, userID primitive.ObjectID, role string) ([]domain.ChatRoom, error) {
	filter := bson.M{}
	if role == "company" {
		filter["employerId"] = userID
	} else {
		filter["seekerId"] = userID
	}

	opts := options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}})
	cursor, err := r.roomsCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []domain.ChatRoom
	if err = cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *mongoChatRepository) SaveMessage(ctx context.Context, msg *domain.ChatMessage) error {
	msg.CreatedAt = time.Now()
	if msg.ID.IsZero() {
		msg.ID = primitive.NewObjectID()
	}
	if msg.Status == "" {
		msg.Status = "sent"
	}
	_, err := r.messagesCol.InsertOne(ctx, msg)
	if err != nil {
		return err
	}

	// Update the ChatRoom's updatedAt time
	_, err = r.roomsCol.UpdateOne(ctx, bson.M{"_id": msg.RoomID}, bson.M{
		"$set": bson.M{"updatedAt": time.Now()},
	})
	return err
}

func (r *mongoChatRepository) GetMessagesByRoom(ctx context.Context, roomID primitive.ObjectID, limit, offset int64) ([]domain.ChatMessage, error) {
	filter := bson.M{"roomId": roomID}
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.messagesCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var msgs []domain.ChatMessage
	if err = cursor.All(ctx, &msgs); err != nil {
		return nil, err
	}

	// Reverse slice so they are in chronological order (oldest first)
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (r *mongoChatRepository) MarkMessagesAsRead(ctx context.Context, roomID, readerID primitive.ObjectID) error {
	_, err := r.messagesCol.UpdateMany(ctx, bson.M{
		"roomId":   roomID,
		"senderId": bson.M{"$ne": readerID},
		"status":   bson.M{"$ne": "seen"},
	}, bson.M{
		"$set": bson.M{"isRead": true, "status": "seen"},
	})
	return err
}

func (r *mongoChatRepository) MarkAllMessagesAsDelivered(ctx context.Context, recipientID primitive.ObjectID) ([]primitive.ObjectID, error) {
	// Find all rooms B is a participant of
	filter := bson.M{
		"$or": []bson.M{
			{"seekerId": recipientID},
			{"employerId": recipientID},
		},
	}
	cursor, err := r.roomsCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []domain.ChatRoom
	if err = cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	roomIDs := make([]primitive.ObjectID, 0, len(rooms))
	for _, room := range rooms {
		roomIDs = append(roomIDs, room.ID)
	}

	if len(roomIDs) == 0 {
		return roomIDs, nil
	}

	// Find all unique room IDs with unsent/delivered messages sent to this recipient
	distinctRoomIDs, err := r.messagesCol.Distinct(ctx, "roomId", bson.M{
		"roomId":   bson.M{"$in": roomIDs},
		"senderId": bson.M{"$ne": recipientID},
		"status":   bson.M{"$in": []interface{}{"sent", "", nil}},
	})
	if err != nil {
		return nil, err
	}

	var affectedRoomIDs []primitive.ObjectID
	for _, id := range distinctRoomIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			affectedRoomIDs = append(affectedRoomIDs, oid)
		}
	}

	if len(affectedRoomIDs) == 0 {
		return affectedRoomIDs, nil
	}

	// Update all "sent" messages in B's rooms where sender is not B to "delivered"
	_, err = r.messagesCol.UpdateMany(ctx, bson.M{
		"roomId":   bson.M{"$in": affectedRoomIDs},
		"senderId": bson.M{"$ne": recipientID},
		"status":   bson.M{"$in": []interface{}{"sent", "", nil}}, // match sent or unset status
	}, bson.M{
		"$set": bson.M{"status": "delivered"},
	})
	if err != nil {
		return nil, err
	}

	return affectedRoomIDs, nil
}

func (r *mongoChatRepository) GetLastMessage(ctx context.Context, roomID primitive.ObjectID) (*domain.ChatMessage, error) {
	var msg domain.ChatMessage
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	err := r.messagesCol.FindOne(ctx, bson.M{"roomId": roomID}, opts).Decode(&msg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *mongoChatRepository) GetUnreadCount(ctx context.Context, roomID, userID primitive.ObjectID) (int64, error) {
	count, err := r.messagesCol.CountDocuments(ctx, bson.M{
		"roomId":   roomID,
		"senderId": bson.M{"$ne": userID},
		"isRead":   false,
	})
	return count, err
}
