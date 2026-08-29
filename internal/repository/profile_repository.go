package repository

import (
	"context"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *domain.SeekerProfile) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.SeekerProfile, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.SeekerProfile, error)
	Update(ctx context.Context, profile *domain.SeekerProfile) error
	FindAllWithAlerts(ctx context.Context) ([]*domain.SeekerProfile, error)
}

type mongoProfileRepository struct {
	collection *mongo.Collection
}

func NewProfileRepository(db *mongo.Database) ProfileRepository {
	return &mongoProfileRepository{
		collection: db.Collection("seeker_profiles"),
	}
}

func (r *mongoProfileRepository) Create(ctx context.Context, profile *domain.SeekerProfile) error {
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	if profile.ID.IsZero() {
		profile.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, profile)
	return err
}

func (r *mongoProfileRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.SeekerProfile, error) {
	var profile domain.SeekerProfile
	err := r.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *mongoProfileRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.SeekerProfile, error) {
	var profile domain.SeekerProfile
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *mongoProfileRepository) Update(ctx context.Context, profile *domain.SeekerProfile) error {
	profile.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": profile.ID}, profile)
	return err
}

func (r *mongoProfileRepository) FindAllWithAlerts(ctx context.Context) ([]*domain.SeekerProfile, error) {
	profiles := make([]*domain.SeekerProfile, 0)
	filter := bson.M{"jobAlertsEnabled": bson.M{"$ne": false}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var profile domain.SeekerProfile
		if err := cursor.Decode(&profile); err != nil {
			return nil, err
		}
		profiles = append(profiles, &profile)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

