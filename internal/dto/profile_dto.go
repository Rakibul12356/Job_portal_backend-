package dto

type UpdateProfileDTO struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Title     string   `json:"title"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Country   string   `json:"country"`
	Zipcode   string   `json:"zipcode"`
	Bio       string   `json:"bio"`
	Skills    []string `json:"skills"`
	Linkedin  string   `json:"linkedin"`
	Github    string   `json:"github"`
	Portfolio string   `json:"portfolio"`
}

type ExperienceDTO struct {
	Company     string `json:"company" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Location    string `json:"location"`
	StartDate   string `json:"startDate" binding:"required"`
	EndDate     string `json:"endDate"`
	Current     bool   `json:"current"`
	Description string `json:"description"`
}

type EducationDTO struct {
	School       string `json:"school" binding:"required"`
	Degree       string `json:"degree" binding:"required"`
	FieldOfStudy string `json:"fieldOfStudy" binding:"required"`
	StartDate    string `json:"startDate" binding:"required"`
	EndDate      string `json:"endDate"`
	Description  string `json:"description"`
}
