package repository

import (
	"context"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Job, error)
	Update(ctx context.Context, job *domain.Job) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context, filter bson.M, sort string, skip, limit int) ([]domain.Job, int64, error)
	FindSimilar(ctx context.Context, job *domain.Job, limit int) ([]domain.Job, error)
	BulkUpdateStatus(ctx context.Context, ids []primitive.ObjectID, companyID primitive.ObjectID, status string) (int64, error)
	BulkDelete(ctx context.Context, ids []primitive.ObjectID, companyID primitive.ObjectID) (int64, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
}

type mongoJobRepository struct {
	collection *mongo.Collection
}

func NewJobRepository(db *mongo.Database) JobRepository {
	return &mongoJobRepository{
		collection: db.Collection("jobs"),
	}
}

func (r *mongoJobRepository) Create(ctx context.Context, job *domain.Job) error {
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, job)
	return err
}

func (r *mongoJobRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Job, error) {
	var job domain.Job
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *mongoJobRepository) Update(ctx context.Context, job *domain.Job) error {
	job.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": job.ID}, job)
	return err
}

func (r *mongoJobRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoJobRepository) FindAll(ctx context.Context, filter bson.M, sort string, skip, limit int) ([]domain.Job, int64, error) {
	// Options
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Sorting
	switch sort {
	case "salary_desc":
		findOptions.SetSort(bson.D{{Key: "salaryMax", Value: -1}, {Key: "createdAt", Value: -1}})
	case "salary_asc":
		findOptions.SetSort(bson.D{{Key: "salaryMin", Value: 1}, {Key: "createdAt", Value: -1}})
	case "oldest":
		findOptions.SetSort(bson.D{{Key: "createdAt", Value: 1}})
	case "newest":
		fallthrough
	default:
		findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var jobs []domain.Job
	if err = cursor.All(ctx, &jobs); err != nil {
		return nil, 0, err
	}
	if jobs == nil {
		jobs = []domain.Job{}
	}

	return jobs, total, nil
}

func (r *mongoJobRepository) FindSimilar(ctx context.Context, job *domain.Job, limit int) ([]domain.Job, error) {
	// Find jobs of similar category/skills, excluding current job, status active
	filter := bson.M{
		"_id":    bson.M{"$ne": job.ID},
		"status": domain.JobStatusActive,
		"$or": []bson.M{
			{"category": job.Category},
			{"skills": bson.M{"$in": job.Skills}},
		},
	}

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []domain.Job
	if err = cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	if jobs == nil {
		jobs = []domain.Job{}
	}

	return jobs, nil
}

func (r *mongoJobRepository) BulkUpdateStatus(ctx context.Context, ids []primitive.ObjectID, companyID primitive.ObjectID, status string) (int64, error) {
	filter := bson.M{
		"_id":       bson.M{"$in": ids},
		"companyId": companyID,
	}
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}
	res, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (r *mongoJobRepository) BulkDelete(ctx context.Context, ids []primitive.ObjectID, companyID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"_id":       bson.M{"$in": ids},
		"companyId": companyID,
	}
	res, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (r *mongoJobRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}
