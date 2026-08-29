package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Email        string              `bson:"email" json:"email"`
	PasswordHash string              `bson:"passwordHash" json:"-"`
	Role         string              `bson:"role" json:"role"` // user | company
	Name         string              `bson:"name" json:"name"`
	FirstName    string              `bson:"firstName" json:"firstName"`
	CompanyID    *primitive.ObjectID `bson:"companyId,omitempty" json:"companyId,omitempty"`
	IsActive          bool                `bson:"isActive" json:"isActive"`
	ResetOTPCode      string              `bson:"resetOtpCode,omitempty" json:"-"`
	ResetOTPExpiresAt time.Time           `bson:"resetOtpExpiresAt,omitempty" json:"-"`
	CreatedAt         time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time           `bson:"updatedAt" json:"updatedAt"`
}
