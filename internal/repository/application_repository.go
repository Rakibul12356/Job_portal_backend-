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

type ApplicationRepository interface {
	Create(ctx context.Context, app *domain.Application) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Application, error)
	Update(ctx context.Context, app *domain.Application) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID, filter bson.M, skip, limit int) ([]domain.Application, int64, error)
	FindByCompanyID(ctx context.Context, companyID primitive.ObjectID, filter bson.M, skip, limit int) ([]domain.Application, int64, error)
	FindByJobIDAndUserID(ctx context.Context, jobID, userID primitive.ObjectID) (*domain.Application, error)
	Count(ctx context.Context, filter bson.M) (int64, error)
}

type mongoApplicationRepository struct {
	collection *mongo.Collection
}

func NewApplicationRepository(db *mongo.Database) ApplicationRepository {
	return &mongoApplicationRepository{
		collection: db.Collection("applications"),
	}
}

func (r *mongoApplicationRepository) Create(ctx context.Context, app *domain.Application) error {
	app.AppliedAt = time.Now()
	app.UpdatedAt = time.Now()
	if app.ID.IsZero() {
		app.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, app)
	return err
}

func (r *mongoApplicationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Application, error) {
	var app domain.Application
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *mongoApplicationRepository) Update(ctx context.Context, app *domain.Application) error {
	app.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": app.ID}, app)
	return err
}

func (r *mongoApplicationRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, filter bson.M, skip, limit int) ([]domain.Application, int64, error) {
	filter["userId"] = userID

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "appliedAt", Value: -1}})

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var apps []domain.Application
	if err = cursor.All(ctx, &apps); err != nil {
		return nil, 0, err
	}
	if apps == nil {
		apps = []domain.Application{}
	}
	return apps, total, nil
}

func (r *mongoApplicationRepository) FindByCompanyID(ctx context.Context, companyID primitive.ObjectID, filter bson.M, skip, limit int) ([]domain.Application, int64, error) {
	filter["companyId"] = companyID

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "appliedAt", Value: -1}})

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var apps []domain.Application
	if err = cursor.All(ctx, &apps); err != nil {
		return nil, 0, err
	}
	if apps == nil {
		apps = []domain.Application{}
	}
	return apps, total, nil
}

func (r *mongoApplicationRepository) FindByJobIDAndUserID(ctx context.Context, jobID, userID primitive.ObjectID) (*domain.Application, error) {
	var app domain.Application
	err := r.collection.FindOne(ctx, bson.M{"jobId": jobID, "userId": userID}).Decode(&app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *mongoApplicationRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}
