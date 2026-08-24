package dto

type SeekerDashboardStats struct {
	TotalApplications int64 `json:"totalApplications"`
	Shortlisted       int64 `json:"shortlisted"`
	Rejected          int64 `json:"rejected"`
	PendingReviews    int64 `json:"pendingReviews"`
	SavedJobs         int64 `json:"savedJobs"`
}

type SeekerDashboardResponseDTO struct {
	Stats           SeekerDashboardStats           `json:"stats"`
	RecentApplied   []SeekerApplicationResponseDTO `json:"recentApplied"`
	RecommendedJobs []JobResponseDTO               `json:"recommendedJobs"`
}

type CompanyDashboardStats struct {
	ActiveJobs     int64 `json:"activeJobs"`
	TotalApplicants int64 `json:"totalApplicants"`
	PendingReviews  int64 `json:"pendingReviews"`
	Shortlisted    int64 `json:"shortlisted"`
}

type CompanyDashboardResponseDTO struct {
	Stats            CompanyDashboardStats         `json:"stats"`
	RecentJobs       []JobResponseDTO              `json:"recentJobs"`
	RecentApplicants []CompanyApplicantResponseDTO `json:"recentApplicants"`
}
