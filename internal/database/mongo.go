package database

import (
	"context"
	"log"
	"time"

	"github.com/rakib/job-portal-api/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var DB *mongo.Database

func ConnectDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB Atlas: %v", err)
	}

	// Ping database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB Atlas ping failed: %v", err)
	}

	log.Println("Successfully connected to MongoDB Atlas!")
	MongoClient = client
	DB = client.Database(cfg.MongoDB)
	return DB
}

func DisconnectDB() {
	if MongoClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := MongoClient.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	} else {
		log.Println("MongoDB Atlas connection closed gracefully.")
	}
}
