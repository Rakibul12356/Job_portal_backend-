package main

import (
	"context"
	"log"
	"time"

	"github.com/rakib/job-portal-api/internal/config"
	"github.com/rakib/job-portal-api/internal/database"
	"github.com/rakib/job-portal-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("Starting database seeding...")

	// 1. Load configuration and connect to database
	cfg := config.LoadConfig()
	db := database.ConnectDB()
	defer database.DisconnectDB()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 2. Clear collections
	clearCollections(ctx, db)

	// 3. Hash passwords
	pwdHash, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	if err != nil {
		log.Fatalf("Failed to hash seed passwords: %v", err)
	}
	hashStr := string(pwdHash)

	// 4. Create Seeker User
	seekerUserID := primitive.NewObjectID()
	seekerUser := domain.User{
		ID:           seekerUserID,
		Email:        "shanjidahmed66@gmail.com",
		PasswordHash: hashStr,
		Role:         domain.RoleUser,
		Name:         "John Doe",
		FirstName:    "John",
		IsActive:     true,
		CreatedAt:    time.Now().Add(-48 * time.Hour),
		UpdatedAt:    time.Now().Add(-48 * time.Hour),
	}

	// 5. Create Employer User & Company
	employerUserID := primitive.NewObjectID()
	companyID := primitive.NewObjectID()
	employerUser := domain.User{
		ID:           employerUserID,
		Email:        "mdrakibulhasan12346@gmail.com",
		PasswordHash: hashStr,
		Role:         domain.RoleCompany,
		Name:         "TechCorp Solutions",
		FirstName:    "Employer",
		CompanyID:    &companyID,
		IsActive:     true,
		CreatedAt:    time.Now().Add(-48 * time.Hour),
		UpdatedAt:    time.Now().Add(-48 * time.Hour),
	}

	company := domain.Company{
		ID:          companyID,
		OwnerUserID: employerUserID,
		Name:        "TechCorp Solutions",
		Industry:    "Information Technology",
		Website:     "https://techcorp.example.com",
		Size:        domain.CompanySize500,
		Type:        domain.CompanyTypePrivate,
		Founded:     "2015",
		About:       "TechCorp Solutions is a leading provider of enterprise cloud computing services. We build scalable SaaS platforms and help companies around the globe digitize their workflows with high quality software engineering architectures.",
		Location: domain.CompanyLocation{
			City:    "San Francisco",
			State:   "California",
			Country: "United States",
		},
		Contact: domain.CompanyContact{
			Phone:        "+1 (555) 123-4567",
			HREmail:      "hr@techcorp.example.com",
			SupportEmail: "support@techcorp.example.com",
		},
		Social: domain.CompanySocial{
			Linkedin:  "https://linkedin.com/company/techcorp",
			Twitter:   "https://twitter.com/techcorp",
			Facebook:  "https://facebook.com/techcorp",
			Instagram: "https://instagram.com/techcorp",
			Github:    "https://github.com/techcorp",
		},
		LogoURL:    cfg.AppBaseURL + "/uploads/logos/default_logo.png",
		Membership: "premium",
		CreatedAt:  time.Now().Add(-48 * time.Hour),
		UpdatedAt:  time.Now().Add(-48 * time.Hour),
	}

	// Save Users and Company
	_, err = db.Collection("users").InsertMany(ctx, []interface{}{seekerUser, employerUser})
	if err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}
	_, err = db.Collection("companies").InsertOne(ctx, company)
	if err != nil {
		log.Fatalf("Failed to seed company: %v", err)
	}

	// 6. Create Seeker Profile
	seekerProfile := domain.SeekerProfile{
		ID:     primitive.NewObjectID(),
		UserID: seekerUserID,
		Title:  "Senior Full Stack Developer",
		Phone:  "+1 (415) 555-0123",
		Bio:    "Experienced full stack software engineer with 6+ years designing web platforms. Passionate about Go, React, and Mongo databases.",
		Location: domain.SeekerLocation{
			City:    "San Francisco",
			State:   "California",
			Country: "United States",
			Zipcode: "94102",
		},
		Skills: []string{"Go", "MongoDB", "JavaScript", "React", "Node.js", "Docker", "REST API"},
		Experience: []domain.Experience{
			{
				ID:          primitive.NewObjectID().Hex(),
				Company:     "Software Dynamics",
				Title:       "Software Developer",
				Location:    "Los Angeles, CA",
				StartDate:   "2021-06",
				EndDate:     "2024-05",
				Current:     false,
				Description: "Designed microservices in Go and built web apps using React. Increased API speed by 40% with database indexing and query tuning.",
			},
			{
				ID:          primitive.NewObjectID().Hex(),
				Company:     "App Innovations",
				Title:       "Junior Developer",
				Location:    "Remote",
				StartDate:   "2019-01",
				EndDate:     "2021-05",
				Current:     false,
				Description: "Assisted in code deployments and fixed Frontend/Backend bugs. Participated in daily standups and agile design review pipelines.",
			},
		},
		Education: []domain.Education{
			{
				ID:           primitive.NewObjectID().Hex(),
				School:       "University of California, Berkeley",
				Degree:       "BS",
				FieldOfStudy: "Computer Science",
				StartDate:    "2015",
				EndDate:      "2018",
				Description:  "Graduated with honors. Key course projects: Distributed Database Storage and Web Architectures.",
			},
		},
		AvatarURL: cfg.AppBaseURL + "/uploads/avatars/default_avatar.jpg",
		Resume: &domain.ResumeMetadata{
			URL:        cfg.AppBaseURL + "/uploads/resumes/default_resume.pdf",
			Filename:   "john_doe_resume.pdf",
			UploadedAt: time.Now().Add(-24 * time.Hour),
		},
		Social: domain.SeekerSocial{
			Linkedin:  "https://linkedin.com/in/johndoe",
			Github:    "https://github.com/johndoe",
			Portfolio: "https://johndoe.dev",
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	_, err = db.Collection("seeker_profiles").InsertOne(ctx, seekerProfile)
	if err != nil {
		log.Fatalf("Failed to seed seeker profile: %v", err)
	}

	// 7. Seed Jobs (9 jobs)
	jobs := createSeedJobs(companyID)
	_, err = db.Collection("jobs").InsertMany(ctx, jobs)
	if err != nil {
		log.Fatalf("Failed to seed jobs: %v", err)
	}

	// 8. Seed Applications (5 applications)
	// Fetch some seeded jobs to get IDs
	var seededJobs []domain.Job
	cur, err := db.Collection("jobs").Find(ctx, bson.M{})
	if err == nil {
		_ = cur.All(ctx, &seededJobs)
	}

	if len(seededJobs) >= 5 {
		apps := []interface{}{
			domain.Application{
				ID:             primitive.NewObjectID(),
				JobID:          seededJobs[0].ID,
				CompanyID:      companyID,
				UserID:         seekerUserID,
				Status:         domain.AppStatusPending,
				CoverMessage:   "I am highly interested in the role and believe my Go skill fits perfectly.",
				ResumeURL:      seekerProfile.Resume.URL,
				ResumeFilename: seekerProfile.Resume.Filename,
				AppliedAt:      time.Now().Add(-6 * time.Hour),
				UpdatedAt:      time.Now().Add(-6 * time.Hour),
			},
			domain.Application{
				ID:             primitive.NewObjectID(),
				JobID:          seededJobs[1].ID,
				CompanyID:      companyID,
				UserID:         seekerUserID,
				Status:         domain.AppStatusShortlisted,
				CoverMessage:   "I have years of experience with React and Node.js backend integration.",
				ResumeURL:      seekerProfile.Resume.URL,
				ResumeFilename: seekerProfile.Resume.Filename,
				AppliedAt:      time.Now().Add(-24 * time.Hour),
				UpdatedAt:      time.Now().Add(-18 * time.Hour),
			},
			domain.Application{
				ID:             primitive.NewObjectID(),
				JobID:          seededJobs[2].ID,
				CompanyID:      companyID,
				UserID:         seekerUserID,
				Status:         domain.AppStatusInterviewed,
				CoverMessage:   "Excited to discuss our alignment during the schedule.",
				ResumeURL:      seekerProfile.Resume.URL,
				ResumeFilename: seekerProfile.Resume.Filename,
				AppliedAt:      time.Now().Add(-48 * time.Hour),
				UpdatedAt:      time.Now().Add(-12 * time.Hour),
			},
			domain.Application{
				ID:             primitive.NewObjectID(),
				JobID:          seededJobs[3].ID,
				CompanyID:      companyID,
				UserID:         seekerUserID,
				Status:         domain.AppStatusRejected,
				CoverMessage:   "Applying for entry level remote design role.",
				ResumeURL:      seekerProfile.Resume.URL,
				ResumeFilename: seekerProfile.Resume.Filename,
				AppliedAt:      time.Now().Add(-72 * time.Hour),
				UpdatedAt:      time.Now().Add(-48 * time.Hour),
			},
		}

		_, err = db.Collection("applications").InsertMany(ctx, apps)
		if err != nil {
			log.Fatalf("Failed to seed applications: %v", err)
		}

		// Update ApplicantsCount for seeded jobs
		for idx, ap := range apps {
			app := ap.(domain.Application)
			_, _ = db.Collection("jobs").UpdateOne(ctx,
				bson.M{"_id": app.JobID},
				bson.M{"$set": bson.M{"applicantsCount": 1}},
			)
			// Seed Saved Jobs for a couple of jobs
			if idx < 2 {
				savedJob := domain.SavedJob{
					ID:        primitive.NewObjectID(),
					UserID:    seekerUserID,
					JobID:     seededJobs[idx+4].ID, // Bookmark jobs without applications
					CreatedAt: time.Now().Add(-2 * time.Hour),
				}
				_, _ = db.Collection("saved_jobs").InsertOne(ctx, savedJob)
			}
		}
	}

	log.Println("Database seeded successfully with mock data!")
}

func clearCollections(ctx context.Context, db *mongo.Database) {
	collections := []string{"users", "companies", "seeker_profiles", "jobs", "applications", "saved_jobs"}
	for _, colName := range collections {
		err := db.Collection(colName).Drop(ctx)
		if err != nil {
			log.Printf("Warning: Failed to drop collection %s: %v", colName, err)
		}
	}
	log.Println("Cleared all existing collection records.")
}

func createSeedJobs(companyID primitive.ObjectID) []interface{} {
	sal1 := 120000
	sal2 := 160000
	sal3 := 80000
	sal4 := 110000
	sal5 := 45
	sal6 := 65

	now := time.Now()

	jobList := []domain.Job{
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "Senior Go Developer",
			Status:          domain.JobStatusActive,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeRemote,
			Category:        domain.CategoryEngineering,
			ExperienceLevel: domain.ExpLevelSenior,
			Location:        "San Francisco, CA",
			SalaryMin:       &sal1,
			SalaryMax:       &sal2,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "We are looking for a Senior Go Engineer to join our backend infrastructure team. You will build highly concurrent API services, scale messaging pipelines, and tune MongoDB database queries.",
			Requirements:    "5+ years developer experience. Proficient in Go concurrency primitives, REST API microservices, and MongoDB Atlas.",
			Benefits:        "Remote first setup, health/dental, 401(k) matching, and $1500 learning stipend.",
			Skills:          []string{"Go", "MongoDB", "REST API", "Microservices"},
			Vacancies:       2,
			Deadline:        now.AddDate(0, 1, 0),
			ApplicantsCount: 0,
			PublishedAt:     &now,
			CreatedAt:       now.Add(-48 * time.Hour),
			UpdatedAt:       now.Add(-48 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "React Frontend Engineer",
			Status:          domain.JobStatusActive,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeHybrid,
			Category:        domain.CategoryEngineering,
			ExperienceLevel: domain.ExpLevelMid,
			Location:        "San Francisco, CA",
			SalaryMin:       &sal3,
			SalaryMax:       &sal4,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "Join our frontend product division to help build the next-generation candidate matching pipeline UI. Work with React, TypeScript, and TailwindCSS to provide gorgeous fluid transitions.",
			Requirements:    "3+ years experience. Expert in React state management, hooks, TypeScript, and responsive design integrations.",
			Benefits:        "Flexible working days, free daily catered lunches, public transit pass, and gym membership.",
			Skills:          []string{"JavaScript", "React", "TypeScript", "TailwindCSS"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, 2, 0),
			ApplicantsCount: 0,
			PublishedAt:     &now,
			CreatedAt:       now.Add(-24 * time.Hour),
			UpdatedAt:       now.Add(-24 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "Full Stack Developer (Node.js)",
			Status:          domain.JobStatusActive,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeOnSite,
			Category:        domain.CategoryEngineering,
			ExperienceLevel: domain.ExpLevelMid,
			Location:        "San Francisco, CA",
			SalaryMin:       &sal4,
			SalaryMax:       &sal1,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "Seeking a versatile developer to connect frontend dashboard components with Node.js APIs and persistence layers. You will own features from design specs to deployment.",
			Requirements:    "React, Node.js express APIs, Mongo BSON querying, and standard code tests execution familiarity.",
			Benefits:        "Full dental/vision packages, company equity shares, and paid parental leave.",
			Skills:          []string{"Node.js", "React", "MongoDB", "Express"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, 0, 15),
			ApplicantsCount: 0,
			PublishedAt:     &now,
			CreatedAt:       now.Add(-6 * time.Hour),
			UpdatedAt:       now.Add(-6 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "Product Designer (UI/UX)",
			Status:          domain.JobStatusActive,
			JobType:         domain.JobTypePartTime,
			WorkMode:        domain.WorkModeRemote,
			Category:        domain.CategoryDesign,
			ExperienceLevel: domain.ExpLevelSenior,
			Location:        "Remote",
			SalaryMin:       &sal5,
			SalaryMax:       &sal6,
			SalaryPeriod:    domain.SalaryPeriodHourly,
			Description:     "Help us conduct user research and turn feedback into beautiful wireframes and interactive web layouts. You will lead the design patterns of our entire product family.",
			Requirements:    "Figma expertise, prototyping, UX research portfolios, and solid visual hierarchy skills.",
			Benefits:        "Flexible working times, high hourly compensation, and brand new equipment setup support.",
			Skills:          []string{"Figma", "UI Design", "UX Research", "Prototyping"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, 0, 20),
			ApplicantsCount: 0,
			PublishedAt:     &now,
			CreatedAt:       now.Add(-12 * time.Hour),
			UpdatedAt:       now.Add(-12 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "Product Manager (SaaS)",
			Status:          domain.JobStatusDraft,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeRemote,
			Category:        domain.CategoryProduct,
			ExperienceLevel: domain.ExpLevelLead,
			Location:        "San Francisco, CA",
			SalaryMin:       &sal2,
			SalaryMax:       &sal2,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "Lead SaaS dashboard feature prioritization and roadmap planning. Act as bridge between engineering teams and business stakeholders.",
			Requirements:    "Prior SaaS PM experience, agile methodologies, product roadmapping tools.",
			Benefits:        "Flexible hours, full premium health plan.",
			Skills:          []string{"Product Management", "Roadmapping", "Agile", "SaaS"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, 3, 0),
			ApplicantsCount: 0,
			CreatedAt:       now.Add(-2 * time.Hour),
			UpdatedAt:       now.Add(-2 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "Content Marketer",
			Status:          domain.JobStatusDraft,
			JobType:         domain.JobTypeContract,
			WorkMode:        domain.WorkModeRemote,
			Category:        domain.CategoryMarketing,
			ExperienceLevel: domain.ExpLevelMid,
			Location:        "Remote",
			SalaryMin:       &sal5,
			SalaryMax:       &sal5,
			SalaryPeriod:    domain.SalaryPeriodHourly,
			Description:     "Draft engaging technical blog posts and documentation tutorials for developers.",
			Requirements:    "Technical copywriting skills, SEO basics, developer marketing.",
			Benefits:        "Remote project coordination.",
			Skills:          []string{"Copywriting", "SEO", "Developer Marketing"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, 1, 0),
			ApplicantsCount: 0,
			CreatedAt:       now.Add(-1 * time.Hour),
			UpdatedAt:       now.Add(-1 * time.Hour),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "DevOps Engineer (SRE)",
			Status:          domain.JobStatusClosed,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeRemote,
			Category:        domain.CategoryEngineering,
			ExperienceLevel: domain.ExpLevelSenior,
			Location:        "Remote",
			SalaryMin:       &sal1,
			SalaryMax:       &sal2,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "Maintain our cloud deploy pipelines and optimize Kubernetes clusters and databases backups.",
			Requirements:    "AWS, Terraform, Kubernetes, CI/CD pipelines scripting.",
			Benefits:        "Excellent stock plan, unlimited PTO.",
			Skills:          []string{"AWS", "Terraform", "Kubernetes", "DevOps"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, -1, 0),
			ApplicantsCount: 0,
			CreatedAt:       now.AddDate(0, -2, 0),
			UpdatedAt:       now.AddDate(0, -1, 0),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "QA Automation Tester",
			Status:          domain.JobStatusClosed,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeHybrid,
			Category:        domain.CategoryEngineering,
			ExperienceLevel: domain.ExpLevelEntry,
			Location:        "San Francisco, CA",
			SalaryMin:       &sal3,
			SalaryMax:       &sal4,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "Write automation script tests for checking API endpoints and web portal widgets.",
			Requirements:    "Selenium, Postman, writing structured automated test suites.",
			Benefits:        "Standard medical packages.",
			Skills:          []string{"Selenium", "Postman", "QA", "Automation"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, -1, 0),
			ApplicantsCount: 0,
			CreatedAt:       now.AddDate(0, -2, 0),
			UpdatedAt:       now.AddDate(0, -1, 0),
		},
		{
			ID:              primitive.NewObjectID(),
			CompanyID:       companyID,
			Title:           "Mobile Engineer (Flutter)",
			Status:          domain.JobStatusExpiringSoon,
			JobType:         domain.JobTypeFullTime,
			WorkMode:        domain.WorkModeHybrid,
			Category:        domain.CategoryEngineering,
			ExperienceLevel: domain.ExpLevelSenior,
			Location:        "San Francisco, CA",
			SalaryMin:       &sal1,
			SalaryMax:       &sal2,
			SalaryPeriod:    domain.SalaryPeriodYearly,
			Description:     "Optimize and publish our mobile apps on iOS and Android platforms.",
			Requirements:    "Flutter, Dart, publishing to App Store and Google Play.",
			Benefits:        "Health plan, learning allowance.",
			Skills:          []string{"Flutter", "Dart", "iOS", "Android"},
			Vacancies:       1,
			Deadline:        now.AddDate(0, 0, 3), // Expiring in 3 days
			ApplicantsCount: 0,
			PublishedAt:     &now,
			CreatedAt:       now.AddDate(0, 0, -27),
			UpdatedAt:       now.AddDate(0, 0, -27),
		},
	}

	interfaces := make([]interface{}, len(jobList))
	for i, v := range jobList {
		interfaces[i] = v
	}
	return interfaces
}
