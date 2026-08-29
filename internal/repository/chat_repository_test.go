package repository

import (
	"context"
	"os"
	"testing"

	"github.com/rakib/job-portal-api/internal/config"
	"github.com/rakib/job-portal-api/internal/database"
	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestChatRepository(t *testing.T) {
	// Set Cwd to root to load .env correctly
	os.Chdir("../../")

	config.LoadConfig()
	db := database.ConnectDB()
	defer database.DisconnectDB()

	repo := NewChatRepository(db)

	ctx := context.Background()
	roomID := primitive.NewObjectID()
	senderID := primitive.NewObjectID()
	recipientID := primitive.NewObjectID()

	// Clean up after test
	defer func() {
		_, _ = db.Collection("chat_messages").DeleteMany(ctx, bson.M{"roomId": roomID})
		_, _ = db.Collection("chat_rooms").DeleteOne(ctx, bson.M{"_id": roomID})
	}()

	// 1. Create a dummy chat room
	room := &domain.ChatRoom{
		ID:         roomID,
		JobID:      primitive.NewObjectID(),
		SeekerID:   recipientID,
		EmployerID: senderID,
	}
	err := repo.CreateRoom(ctx, room)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// 2. Save message - check default status "sent"
	msg1 := &domain.ChatMessage{
		ID:       primitive.NewObjectID(),
		RoomID:   roomID,
		SenderID: senderID,
		Message:  "Hello seeker",
	}
	err = repo.SaveMessage(ctx, msg1)
	if err != nil {
		t.Fatalf("Failed to save message 1: %v", err)
	}

	// Retrieve message and check status
	var savedMsg1 domain.ChatMessage
	err = db.Collection("chat_messages").FindOne(ctx, bson.M{"_id": msg1.ID}).Decode(&savedMsg1)
	if err != nil {
		t.Fatalf("Failed to retrieve saved message 1: %v", err)
	}
	if savedMsg1.Status != "sent" {
		t.Errorf("Expected status to be 'sent', got %s", savedMsg1.Status)
	}

	// 3. MarkAllMessagesAsDelivered sync
	affectedRooms, err := repo.MarkAllMessagesAsDelivered(ctx, recipientID)
	if err != nil {
		t.Fatalf("Failed to mark messages as delivered: %v", err)
	}
	if len(affectedRooms) != 1 || affectedRooms[0] != roomID {
		t.Errorf("Expected affected rooms to contain room %v, got %v", roomID, affectedRooms)
	}

	// Verify message status is updated to "delivered"
	var savedMsg1Deliv domain.ChatMessage
	err = db.Collection("chat_messages").FindOne(ctx, bson.M{"_id": msg1.ID}).Decode(&savedMsg1Deliv)
	if err != nil {
		t.Fatalf("Failed to retrieve message after delivery sync: %v", err)
	}
	if savedMsg1Deliv.Status != "delivered" {
		t.Errorf("Expected status to be 'delivered', got %s", savedMsg1Deliv.Status)
	}

	// 4. MarkMessagesAsRead (Seen Trigger)
	err = repo.MarkMessagesAsRead(ctx, roomID, recipientID)
	if err != nil {
		t.Fatalf("Failed to mark messages as read: %v", err)
	}

	// Verify message status is updated to "seen"
	var savedMsg1Seen domain.ChatMessage
	err = db.Collection("chat_messages").FindOne(ctx, bson.M{"_id": msg1.ID}).Decode(&savedMsg1Seen)
	if err != nil {
		t.Fatalf("Failed to retrieve message after seen trigger: %v", err)
	}
	if savedMsg1Seen.Status != "seen" || !savedMsg1Seen.IsRead {
		t.Errorf("Expected status to be 'seen' and isRead to be true, got status %s and isRead %v", savedMsg1Seen.Status, savedMsg1Seen.IsRead)
	}

	// 5. Test sender messages are NOT marked as seen by sender
	msg2 := &domain.ChatMessage{
		ID:       primitive.NewObjectID(),
		RoomID:   roomID,
		SenderID: recipientID, // sent by recipient
		Message:  "Reply from seeker",
		Status:   "delivered",
	}
	err = repo.SaveMessage(ctx, msg2)
	if err != nil {
		t.Fatalf("Failed to save message 2: %v", err)
	}

	// Mark messages read by recipient (should not affect msg2 since recipient is sender of msg2)
	err = repo.MarkMessagesAsRead(ctx, roomID, recipientID)
	if err != nil {
		t.Fatalf("Failed to mark messages read: %v", err)
	}

	var savedMsg2 domain.ChatMessage
	err = db.Collection("chat_messages").FindOne(ctx, bson.M{"_id": msg2.ID}).Decode(&savedMsg2)
	if err != nil {
		t.Fatalf("Failed to retrieve message 2: %v", err)
	}
	if savedMsg2.Status != "delivered" {
		t.Errorf("Expected recipient's own message status to remain 'delivered', got %s", savedMsg2.Status)
	}
}
