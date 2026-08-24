package dto

type UpdateCompanySettingsDTO struct {
	CompanyName  string `json:"companyName" binding:"required"`
	Industry     string `json:"industry"`
	CompanySize  string `json:"companySize"`
	CompanyType  string `json:"companyType"`
	Website      string `json:"website"`
	Founded      string `json:"founded"`
	About        string `json:"about"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	HREmail      string `json:"hrEmail"`
	SupportEmail string `json:"supportEmail"`
	Linkedin     string `json:"linkedin"`
	Twitter      string `json:"twitter"`
	Facebook     string `json:"facebook"`
	Instagram    string `json:"instagram"`
	Github       string `json:"github"`
}

type CompanySettingsResponseDTO struct {
	CompanyName  string `json:"companyName"`
	Industry     string `json:"industry"`
	CompanySize  string `json:"companySize"`
	CompanyType  string `json:"companyType"`
	Website      string `json:"website"`
	Founded      string `json:"founded"`
	About        string `json:"about"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	HREmail      string `json:"hrEmail"`
	SupportEmail string `json:"supportEmail"`
	Linkedin     string `json:"linkedin"`
	Twitter      string `json:"twitter"`
	Facebook     string `json:"facebook"`
	Instagram    string `json:"instagram"`
	Github       string `json:"github"`
}

type PublicCompanyProfileResponseDTO struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Industry    string           `json:"industry"`
	Website     string           `json:"website"`
	Size        string           `json:"size"`
	Type        string           `json:"type"`
	Founded     string           `json:"founded"`
	About       string           `json:"about"`
	LogoURL     string           `json:"logoUrl"`
	City        string           `json:"city"`
	State       string           `json:"state"`
	Country     string           `json:"country"`
	Linkedin    string           `json:"linkedin"`
	Twitter     string           `json:"twitter"`
	Facebook    string           `json:"facebook"`
	Instagram   string           `json:"instagram"`
	Github      string           `json:"github"`
	OpenJobs    []JobResponseDTO `json:"openJobs"`
}
