package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompanyLocation struct {
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	Country string `bson:"country" json:"country"`
}

type CompanyContact struct {
	Phone        string `bson:"phone" json:"phone"`
	HREmail      string `bson:"hrEmail" json:"hrEmail"`
	SupportEmail string `bson:"supportEmail" json:"supportEmail"`
}

type CompanySocial struct {
	Linkedin  string `bson:"linkedin" json:"linkedin"`
	Twitter   string `bson:"twitter" json:"twitter"`
	Facebook  string `bson:"facebook" json:"facebook"`
	Instagram string `bson:"instagram" json:"instagram"`
	Github    string `bson:"github" json:"github"`
}

type Company struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerUserID primitive.ObjectID `bson:"ownerUserId" json:"ownerUserId"`
	Name        string             `bson:"name" json:"name"`
	Industry    string             `bson:"industry" json:"industry"`
	Website     string             `bson:"website" json:"website"`
	Size        string             `bson:"size" json:"size"`
	Type        string             `bson:"type" json:"type"`
	Founded     string             `bson:"founded" json:"founded"`
	About       string             `bson:"about" json:"about"`
	Location    CompanyLocation    `bson:"location" json:"location"`
	Contact     CompanyContact     `bson:"contact" json:"contact"`
	Social      CompanySocial      `bson:"social" json:"social"`
	LogoURL     string             `bson:"logoUrl" json:"logoUrl"`
	Membership  string             `bson:"membership" json:"membership"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
