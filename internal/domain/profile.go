package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SeekerLocation struct {
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
	Country string `bson:"country" json:"country"`
	Zipcode string `bson:"zipcode" json:"zipcode"`
}

type SeekerSocial struct {
	Linkedin  string `bson:"linkedin" json:"linkedin"`
	Github    string `bson:"github" json:"github"`
	Portfolio string `bson:"portfolio" json:"portfolio"`
}

type ResumeMetadata struct {
	URL        string    `bson:"url" json:"url"`
	Filename   string    `bson:"filename" json:"filename"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

type Experience struct {
	ID          string    `bson:"id" json:"id"` // Unique string identifier within the list
	Company     string    `bson:"company" json:"company"`
	Title       string    `bson:"title" json:"title"`
	Location    string    `bson:"location" json:"location"`
	StartDate   string    `bson:"startDate" json:"startDate"`
	EndDate     string    `bson:"endDate" json:"endDate"`
	Current     bool      `bson:"current" json:"current"`
	Description string    `bson:"description" json:"description"`
}

type Education struct {
	ID           string    `bson:"id" json:"id"` // Unique string identifier within the list
	School       string    `bson:"school" json:"school"`
	Degree       string    `bson:"degree" json:"degree"`
	FieldOfStudy string    `bson:"fieldOfStudy" json:"fieldOfStudy"`
	StartDate    string    `bson:"startDate" json:"startDate"`
	EndDate      string    `bson:"endDate" json:"endDate"`
	Description  string    `bson:"description" json:"description"`
}

type SeekerProfile struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Title      string             `bson:"title" json:"title"`
	Phone      string             `bson:"phone" json:"phone"`
	Bio        string             `bson:"bio" json:"bio"`
	Location   SeekerLocation     `bson:"location" json:"location"`
	Skills     []string           `bson:"skills" json:"skills"`
	Experience []Experience       `bson:"experience" json:"experience"`
	Education  []Education        `bson:"education" json:"education"`
	AvatarURL  string             `bson:"avatarUrl" json:"avatarUrl"`
	Resume     *ResumeMetadata    `bson:"resume,omitempty" json:"resume,omitempty"`
	Social           SeekerSocial       `bson:"social" json:"social"`
	JobAlertsEnabled bool               `bson:"jobAlertsEnabled" json:"jobAlertsEnabled"`
	CreatedAt        time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time          `bson:"updatedAt" json:"updatedAt"`
}
