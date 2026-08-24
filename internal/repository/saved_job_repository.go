package repository

import (
	"context"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SavedJobRepository interface {
	Create(ctx context.Context, savedJob *domain.SavedJob) error
	Delete(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]domain.SavedJob, error)
	IsSaved(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) (bool, error)
}

type mongoSavedJobRepository struct {
	collection *mongo.Collection
}

func NewSavedJobRepository(db *mongo.Database) SavedJobRepository {
	return &mongoSavedJobRepository{
		collection: db.Collection("saved_jobs"),
	}
}

func (r *mongoSavedJobRepository) Create(ctx context.Context, savedJob *domain.SavedJob) error {
	savedJob.CreatedAt = time.Now()
	if savedJob.ID.IsZero() {
		savedJob.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, savedJob)
	return err
}

func (r *mongoSavedJobRepository) Delete(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"userId": userID,
		"jobId":  jobID,
	})
	return err
}

func (r *mongoSavedJobRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]domain.SavedJob, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var saved []domain.SavedJob
	if err = cursor.All(ctx, &saved); err != nil {
		return nil, err
	}
	if saved == nil {
		saved = []domain.SavedJob{}
	}
	return saved, nil
}

func (r *mongoSavedJobRepository) IsSaved(ctx context.Context, userID primitive.ObjectID, jobID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"userId": userID,
		"jobId":  jobID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
