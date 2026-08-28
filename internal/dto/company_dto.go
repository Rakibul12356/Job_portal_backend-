package dto

type UpdateCompanySettingsDTO struct {
	CompanyName  string `json:"companyName"`
	Name         string `json:"name"` // Alias for companyName
	Industry     string `json:"industry"`
	CompanySize  string `json:"companySize"`
	Size         string `json:"size"` // Alias for companySize
	CompanyType  string `json:"companyType"`
	Type         string `json:"type"` // Alias for companyType
	Website      string `json:"website"`
	Founded      string `json:"founded"`
	About        string `json:"about"`
	Description  string `json:"description"` // Alias for about
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
	ID           string `json:"id,omitempty"`
	CompanyName  string `json:"companyName"`
	Name         string `json:"name,omitempty"`
	AccountEmail string `json:"accountEmail,omitempty"` // Read-only login email
	Industry     string `json:"industry"`
	CompanySize  string `json:"companySize"`
	Size         string `json:"size,omitempty"`
	CompanyType  string `json:"companyType"`
	Type         string `json:"type,omitempty"`
	Website      string `json:"website"`
	Founded      string `json:"founded"`
	About        string `json:"about"`
	LogoURL      string `json:"logoUrl,omitempty"`
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
