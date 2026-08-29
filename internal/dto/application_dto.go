package dto

import "time"

type UpdateApplicationStatusDTO struct {
	Status        string `json:"status" binding:"required,oneof=shortlisted rejected interviewed"`
	InterviewDate string `json:"interviewDate"`
	InterviewTime string `json:"interviewTime"`
	Notes         string `json:"notes"`
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
	InterviewDate  string    `json:"interviewDate,omitempty"`
	InterviewTime  string    `json:"interviewTime,omitempty"`
	InterviewNotes string    `json:"interviewNotes,omitempty"`
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
	InterviewDate  string    `json:"interviewDate,omitempty"`
	InterviewTime  string    `json:"interviewTime,omitempty"`
	InterviewNotes string    `json:"interviewNotes,omitempty"`
}
