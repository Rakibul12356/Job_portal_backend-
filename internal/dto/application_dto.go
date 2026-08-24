package dto

import "time"

type UpdateApplicationStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=shortlisted rejected interviewed"`
}

type SeekerApplicationResponseDTO struct {
	ID             string    `json:"id"`
	JobID          string    `json:"jobId"`
	JobTitle       string    `json:"jobTitle"`
	Company        string    `json:"company"`
	CompanyID      string    `json:"companyId"`
	CompanyLogo    string    `json:"companyLogo"`
	Status         string    `json:"status"`
	AppliedAt      time.Time `json:"appliedAt"`
	ResumeURL      string    `json:"resumeUrl"`
	ResumeFilename string    `json:"resumeFilename"`
	CoverMessage   string    `json:"coverMessage"`
	Location       string    `json:"location"`
	JobType        string    `json:"jobType"`
}

type CompanyApplicantResponseDTO struct {
	ID             string    `json:"id"`
	JobID          string    `json:"jobId"`
	JobTitle       string    `json:"jobTitle"`
	UserID         string    `json:"userId"`
	SeekerName     string    `json:"seekerName"`
	SeekerEmail    string    `json:"seekerEmail"`
	SeekerTitle    string    `json:"seekerTitle"`
	SeekerPhone    string    `json:"seekerPhone"`
	SeekerAvatar   string    `json:"seekerAvatar"`
	SeekerSkills   []string  `json:"seekerSkills"`
	Status         string    `json:"status"`
	AppliedAt      time.Time `json:"appliedAt"`
	ResumeURL      string    `json:"resumeUrl"`
	ResumeFilename string    `json:"resumeFilename"`
	CoverMessage   string    `json:"coverMessage"`
}
