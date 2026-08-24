package repository

import (
	"context"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompanyRepository interface {
	Create(ctx context.Context, company *domain.Company) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Company, error)
	FindByOwnerID(ctx context.Context, ownerID primitive.ObjectID) (*domain.Company, error)
	Update(ctx context.Context, company *domain.Company) error
	FindAll(ctx context.Context) ([]domain.Company, error)
}

type mongoCompanyRepository struct {
	collection *mongo.Collection
}

func NewCompanyRepository(db *mongo.Database) CompanyRepository {
	return &mongoCompanyRepository{
		collection: db.Collection("companies"),
	}
}

func (r *mongoCompanyRepository) Create(ctx context.Context, company *domain.Company) error {
	company.CreatedAt = time.Now()
	company.UpdatedAt = time.Now()
	if company.ID.IsZero() {
		company.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, company)
	return err
}

func (r *mongoCompanyRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Company, error) {
	var company domain.Company
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&company)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *mongoCompanyRepository) FindByOwnerID(ctx context.Context, ownerID primitive.ObjectID) (*domain.Company, error) {
	var company domain.Company
	err := r.collection.FindOne(ctx, bson.M{"ownerUserId": ownerID}).Decode(&company)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *mongoCompanyRepository) Update(ctx context.Context, company *domain.Company) error {
	company.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": company.ID}, company)
	return err
}

func (r *mongoCompanyRepository) FindAll(ctx context.Context) ([]domain.Company, error) {
	var companies []domain.Company
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &companies); err != nil {
		return nil, err
	}
	if companies == nil {
		companies = []domain.Company{}
	}
	return companies, nil
}
