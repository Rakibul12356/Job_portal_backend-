package dto

import "time"

type CreateJobDTO struct {
	Title           string    `json:"title" binding:"required"`
	JobType         string    `json:"jobType" binding:"required"`
	WorkMode        string    `json:"workMode" binding:"required"`
	Category        string    `json:"category" binding:"required"`
	ExperienceLevel string    `json:"experienceLevel" binding:"required"`
	Location        string    `json:"location" binding:"required"`
	SalaryMin       *int      `json:"salaryMin"`
	SalaryMax       *int      `json:"salaryMax"`
	SalaryPeriod    string    `json:"salaryPeriod"`
	Description     string    `json:"description" binding:"required"`
	Requirements    string    `json:"requirements" binding:"required"`
	Benefits        string    `json:"benefits"`
	Skills          []string  `json:"skills" binding:"required,min=1"`
	Vacancies       int       `json:"vacancies" binding:"required,gt=0"`
	Deadline        time.Time `json:"deadline" binding:"required"`
	Status          string    `json:"status"` // draft | active (default active)
}

type UpdateJobDTO struct {
	Title           string    `json:"title"`
	JobType         string    `json:"jobType"`
	WorkMode        string    `json:"workMode"`
	Category        string    `json:"category"`
	ExperienceLevel string    `json:"experienceLevel"`
	Location        string    `json:"location"`
	SalaryMin       *int      `json:"salaryMin"`
	SalaryMax       *int      `json:"salaryMax"`
	SalaryPeriod    string    `json:"salaryPeriod"`
	Description     string    `json:"description"`
	Requirements    string    `json:"requirements"`
	Benefits        string    `json:"benefits"`
	Skills          []string  `json:"skills"`
	Vacancies       int       `json:"vacancies"`
	Deadline        time.Time `json:"deadline"`
	Status          string    `json:"status"`
}

type CompanyInfoDTO struct {
	Name        string `json:"name,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Industry    string `json:"industry"`
	About       string `json:"about"`
	Website     string `json:"website"`
	Location    string `json:"location"`
	Employees   string `json:"employees"`
	CompanySize string `json:"companySize,omitempty"`
	CompanyType string `json:"companyType,omitempty"`
	Founded     string `json:"founded"`
	LogoURL     string `json:"logoUrl"`
}

type JobResponseDTO struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Company         string          `json:"company"`
	CompanyID       string          `json:"companyId"`
	LogoURL         string          `json:"logoUrl,omitempty"`
	Location        string          `json:"location"`
	PostedAt        time.Time       `json:"postedAt"`
	PostedLabel     string          `json:"postedLabel"`
	Category        string          `json:"category"`
	Description     string          `json:"description"`
	Requirements    string          `json:"requirements,omitempty"`
	Benefits        string          `json:"benefits,omitempty"`
	Tags            []string        `json:"tags"`
	Salary          string          `json:"salary"`
	SalaryMin       *int            `json:"salaryMin"`
	SalaryMax       *int            `json:"salaryMax"`
	SalaryPeriod    string          `json:"salaryPeriod"`
	Applicants      int             `json:"applicants"`
	JobType         string          `json:"jobType"`
	WorkMode        string          `json:"workMode"`
	ExperienceLevel string          `json:"experienceLevel"`
	Deadline        time.Time       `json:"deadline"`
	Status          string          `json:"status"`
	Skills          []string        `json:"skills"`
	Vacancies       int             `json:"vacancies"`
	CompanyInfo     *CompanyInfoDTO `json:"companyInfo,omitempty"`
}

type BulkJobActionDTO struct {
	JobIDs []string `json:"jobIds" binding:"required,min=1"`
	Action string   `json:"action" binding:"required,oneof=activate deactivate delete"`
}
