package domain

// Roles
const (
	RoleUser    = "user"    // job seeker
	RoleCompany = "company" // employer
)

// Job status
const (
	JobStatusDraft        = "draft"
	JobStatusActive       = "active"
	JobStatusExpiringSoon = "expiring_soon"
	JobStatusClosed       = "closed"
)

// Application status
const (
	AppStatusPending     = "pending"
	AppStatusShortlisted = "shortlisted"
	AppStatusInterviewed = "interviewed"
	AppStatusRejected    = "rejected"
	AppStatusWithdrawn   = "withdrawn"
)

// Job types
const (
	JobTypeFullTime   = "full-time"
	JobTypePartTime   = "part-time"
	JobTypeContract   = "contract"
	JobTypeFreelance  = "freelance"
	JobTypeInternship = "internship"
)

// Work modes
const (
	WorkModeOnSite = "on-site"
	WorkModeRemote = "remote"
	WorkModeHybrid = "hybrid"
)

// Experience levels
const (
	ExpLevelEntry  = "entry"
	ExpLevelMid    = "mid"
	ExpLevelSenior = "senior"
	ExpLevelLead   = "lead"
	ExpLevelExpert = "expert"
)

// Categories
const (
	CategoryEngineering = "engineering"
	CategoryDesign      = "design"
	CategoryProduct     = "product"
	CategoryMarketing   = "marketing"
	CategorySales       = "sales"
	CategoryHR          = "hr"
	CategoryFinance     = "finance"
	CategoryOther       = "other"
)

// Salary periods
const (
	SalaryPeriodYearly  = "yearly"
	SalaryPeriodMonthly = "monthly"
	SalaryPeriodHourly  = "hourly"
)

// Company sizes
const (
	CompanySize1_10    = "1-10"
	CompanySize50      = "50"
	CompanySize200     = "200"
	CompanySize500     = "500"
	CompanySize1000    = "1000"
	CompanySize5000    = "5000"
	CompanySize10000   = "10000"
	CompanySize10001Plus = "10001+"
)

// Company types
const (
	CompanyTypeStartup      = "startup"
	CompanyTypePrivate      = "private"
	CompanyTypePublic       = "public"
	CompanyTypeNonProfit    = "non-profit"
	CompanyTypeGovernment   = "government"
	CompanyTypeEducational  = "educational"
	CompanyTypeSelfEmployed = "self-employed"
	CompanyTypePartnership  = "partnership"
)
