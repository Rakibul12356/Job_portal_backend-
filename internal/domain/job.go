package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobSalary struct {
	Min    *int   `bson:"min,omitempty" json:"min,omitempty"`
	Max    *int   `bson:"max,omitempty" json:"max,omitempty"`
	Period string `bson:"period" json:"period"`
}

type Job struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CompanyID       primitive.ObjectID `bson:"companyId" json:"companyId"`
	Title           string             `bson:"title" json:"title"`
	Status          string             `bson:"status" json:"status"` // draft | active | expiring_soon | closed
	JobType         string             `bson:"jobType" json:"jobType"`
	WorkMode        string             `bson:"workMode" json:"workMode"`
	Category        string             `bson:"category" json:"category"`
	ExperienceLevel string             `bson:"experienceLevel" json:"experienceLevel"`
	Location        string             `bson:"location" json:"location"`
	SalaryMin       *int               `bson:"salaryMin,omitempty" json:"salaryMin,omitempty"`
	SalaryMax       *int               `bson:"salaryMax,omitempty" json:"salaryMax,omitempty"`
	SalaryPeriod    string             `bson:"salaryPeriod" json:"salaryPeriod"`
	Description     string             `bson:"description" json:"description"`
	Requirements    string             `bson:"requirements" json:"requirements"`
	Benefits        string             `bson:"benefits" json:"benefits"`
	Skills          []string           `bson:"skills" json:"skills"`
	Vacancies       int                `bson:"vacancies" json:"vacancies"`
	Deadline        time.Time          `bson:"deadline" json:"deadline"`
	ApplicantsCount int                `bson:"applicantsCount" json:"applicantsCount"`
	PublishedAt     *time.Time         `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}
