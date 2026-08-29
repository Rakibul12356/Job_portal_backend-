package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MigrateIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Users collection indexes
	usersCol := db.Collection("users")
	_, err := usersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Failed to create users email index: %v", err)
	}

	// 2. Companies indexes
	companiesCol := db.Collection("companies")
	_, err = companiesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "ownerUserId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Failed to create companies ownerUserId index: %v", err)
	}
	_, err = companiesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: "text"},
			{Key: "industry", Value: "text"},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create companies text index: %v", err)
	}

	// 3. Jobs indexes
	jobsCol := db.Collection("jobs")
	_, err = jobsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "companyId", Value: 1},
			{Key: "status", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create jobs companyId-status-createdAt index: %v", err)
	}
	_, err = jobsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create jobs status-createdAt index: %v", err)
	}
	_, err = jobsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "description", Value: "text"},
			{Key: "skills", Value: "text"},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create jobs text index: %v", err)
	}
	_, err = jobsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "salaryMin", Value: 1},
			{Key: "salaryMax", Value: 1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create jobs salary index: %v", err)
	}
	_, err = jobsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "deadline", Value: 1}},
	})
	if err != nil {
		log.Printf("Warning: Failed to create jobs deadline index: %v", err)
	}

	// 4. Applications indexes
	appsCol := db.Collection("applications")
	_, err = appsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "appliedAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create applications userId index: %v", err)
	}
	_, err = appsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "companyId", Value: 1},
			{Key: "status", Value: 1},
			{Key: "appliedAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create applications companyId index: %v", err)
	}
	_, err = appsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "jobId", Value: 1},
			{Key: "status", Value: 1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create applications jobId status index: %v", err)
	}
	_, err = appsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "jobId", Value: 1},
			{Key: "userId", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Failed to create applications unique jobId-userId index: %v", err)
	}

	// 5. Saved Jobs indexes
	savedCol := db.Collection("saved_jobs")
	_, err = savedCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "jobId", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Failed to create saved_jobs unique userId-jobId index: %v", err)
	}
	_, err = savedCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create saved_jobs userId-createdAt index: %v", err)
	}

	// 6. Chat Rooms indexes
	roomsCol := db.Collection("chat_rooms")
	_, err = roomsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "jobId", Value: 1},
			{Key: "seekerId", Value: 1},
			{Key: "employerId", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Failed to create chat_rooms unique participants index: %v", err)
	}
	_, err = roomsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "seekerId", Value: 1}},
	})
	if err != nil {
		log.Printf("Warning: Failed to create chat_rooms seekerId index: %v", err)
	}
	_, err = roomsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "employerId", Value: 1}},
	})
	if err != nil {
		log.Printf("Warning: Failed to create chat_rooms employerId index: %v", err)
	}

	// 7. Chat Messages indexes
	messagesCol := db.Collection("chat_messages")
	_, err = messagesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "roomId", Value: 1},
			{Key: "createdAt", Value: -1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create chat_messages roomId-createdAt index: %v", err)
	}
	_, err = messagesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "roomId", Value: 1},
			{Key: "senderId", Value: 1},
			{Key: "status", Value: 1},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create chat_messages roomId-senderId-status index: %v", err)
	}

	log.Println("MongoDB database indexes migrated successfully.")
}
