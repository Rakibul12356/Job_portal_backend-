package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JobID          primitive.ObjectID `bson:"jobId" json:"jobId"`
	CompanyID      primitive.ObjectID `bson:"companyId" json:"companyId"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId"`
	Status         string             `bson:"status" json:"status"` // pending | shortlisted | interviewed | rejected | withdrawn
	CoverMessage   string             `bson:"coverMessage" json:"coverMessage"`
	ResumeURL      string             `bson:"resumeUrl" json:"resumeUrl"`
	ResumeFilename string             `bson:"resumeFilename" json:"resumeFilename"`
	AppliedAt      time.Time          `bson:"appliedAt" json:"appliedAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	ReviewedAt     *time.Time         `bson:"reviewedAt,omitempty" json:"reviewedAt,omitempty"`
}
