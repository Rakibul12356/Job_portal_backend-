package dto

type RegisterSeekerDTO struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Phone           string `json:"phone"`
	Experience      string `json:"experience"` // entry, mid, senior
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	TermsAccepted   bool   `json:"termsAccepted" binding:"required,eq=true"`
}

type RegisterEmployerDTO struct {
	CompanyName        string `json:"companyName" binding:"required"`
	Email              string `json:"email" binding:"required,email"`
	Website            string `json:"website"`
	Industry           string `json:"industry"`
	CompanySize        string `json:"companySize"`
	FoundedYear        int    `json:"foundedYear"`
	Location           string `json:"location"`
	Description        string `json:"description" binding:"required,min=100"`
	Password           string `json:"password" binding:"required,min=8"`
	ConfirmPassword    string `json:"confirmPassword" binding:"required,eqfield=Password"`
	TermsAccepted      bool   `json:"termsAccepted" binding:"required,eq=true"`
	VerifiedAuthorized bool   `json:"verifiedAuthorized" binding:"required,eq=true"`
	MarketingOptIn     bool   `json:"marketingOptIn"`
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponseDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	FirstName string  `json:"firstName"`
	Role      string  `json:"role"`
	CompanyID *string `json:"companyId"`
}

type LoginResponseDTO struct {
	AccessToken  string          `json:"accessToken"`
	RefreshToken string          `json:"refreshToken"`
	User         UserResponseDTO `json:"user"`
}

type RefreshDTO struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ForgotPasswordDTO struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordDTO struct {
	Email           string `json:"email" binding:"required,email"`
	OTP             string `json:"otp" binding:"required,len=6"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

