package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rakib/job-portal-api/internal/domain"
	"github.com/rakib/job-portal-api/internal/pkg/utils"
	"github.com/rakib/job-portal-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

// StartJobAlertsWorker initializes and runs the background job alerts scan on a ticker.
func StartJobAlertsWorker(
	ctx context.Context,
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	jobRepo repository.JobRepository,
	companyRepo repository.CompanyRepository,
) {
	log.Println("Initializing native Go weekly Job Alerts background worker...")

	// Default ticker interval is 7 days (Weekly)
	interval := 7 * 24 * time.Hour

	// Switch to 2 minutes in development mode for easy testing/demoing
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	if env == "development" {
		interval = 2 * time.Minute
		log.Println("Worker is running in development mode (Ticker interval: 2 minutes)")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run an initial scan immediately on server boot
	runJobAlertsScan(ctx, userRepo, profileRepo, jobRepo, companyRepo, env)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Job Alerts background worker gracefully...")
			return
		case <-ticker.C:
			runJobAlertsScan(ctx, userRepo, profileRepo, jobRepo, companyRepo, env)
		}
	}
}

func runJobAlertsScan(
	ctx context.Context,
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	jobRepo repository.JobRepository,
	companyRepo repository.CompanyRepository,
	env string,
) {
	log.Println("Background Job Alert scan triggered...")

	// 1. Fetch all profiles where alerts are enabled
	profiles, err := profileRepo.FindAllWithAlerts(ctx)
	if err != nil {
		log.Printf("Worker error finding profiles: %v", err)
		return
	}

	log.Printf("Found %d seeker profiles with job alerts enabled.", len(profiles))

	// Scan window matches the ticker interval (7 days for prod, 5 minutes for dev to capture recent entries)
	scanWindow := 7 * 24 * time.Hour
	if env == "development" {
		scanWindow = 5 * time.Minute
	}

	for _, profile := range profiles {
		if len(profile.Skills) == 0 {
			continue // Skip users who have no skills registered
		}

		// 2. Query MongoDB for active jobs created within the scan window matching user skills
		filter := bson.M{
			"status":    domain.JobStatusActive,
			"createdAt": bson.M{"$gte": time.Now().Add(-scanWindow)},
			"skills":    bson.M{"$in": profile.Skills},
		}

		jobs, _, err := jobRepo.FindAll(ctx, filter, "newest", 0, 5) // Retrieve top 5 newest jobs
		if err != nil {
			log.Printf("Worker error querying jobs for user %s: %v", profile.UserID.Hex(), err)
			continue
		}

		if len(jobs) == 0 {
			continue // No new matching jobs
		}

		// 3. Fetch user email and name details
		user, err := userRepo.FindByID(ctx, profile.UserID)
		if err != nil || user == nil {
			continue
		}

		log.Printf("Sending job alert digest to %s containing %d jobs.", user.Email, len(jobs))

		// 4. Generate HTML job digest list
		var jobsListHTML string
		for _, job := range jobs {
			// Fetch company details for company name
			company, err := companyRepo.FindByID(ctx, job.CompanyID)
			companyName := "Employer"
			if err == nil && company != nil {
				companyName = company.Name
			}

			jobsListHTML += fmt.Sprintf(`
				<div style="padding: 15px; border-bottom: 1px solid #e2e8f0;">
					<h3 style="margin: 0 0 5px 0; color: #0f172a; font-size: 16px;">%s</h3>
					<p style="margin: 0 0 5px 0; color: #475569; font-size: 14px;"><strong>Company:</strong> %s | <strong>Location:</strong> %s</p>
					<p style="margin: 0 0 10px 0; color: #64748b; font-size: 13px;"><strong>Skills Required:</strong> %s</p>
					<a href="http://localhost:5174/jobs/%s" style="background-color: #2563eb; color: white; padding: 6px 12px; text-decoration: none; border-radius: 4px; font-size: 13px; font-weight: bold; display: inline-block;">View Opening</a>
				</div>
			`, job.Title, companyName, job.Location, formatSkillsList(job.Skills), job.ID.Hex())
		}

		emailBody := fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; padding: 20px; border: 1px solid #eee; border-radius: 8px; max-width: 600px; margin: 0 auto; line-height: 1.5;">
				<h2 style="color: #2563eb; text-align: center;">Weekly Job Match Alerts</h2>
				<p>Hello %s,</p>
				<p>We found new job opportunities matching your profile skills (<strong>%s</strong>):</p>
				<div style="margin: 20px 0; border: 1px solid #e2e8f0; border-radius: 6px; overflow: hidden;">
					%s
				</div>
				<p style="font-size: 12px; color: #64748b;">You received this because you enabled Job Alerts. You can disable this anytime from your profile settings.</p>
				<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;" />
				<p style="font-size: 11px; color: #94a3b8; text-align: center;">Online Job Portal &copy; 2026. All rights reserved.</p>
			</div>
		`, user.Name, formatSkillsList(profile.Skills), jobsListHTML)

		// 5. Send email asynchronously
		go func(email, body string) {
			_ = utils.SendEmail(email, "New Job Match Alerts For You", body)
		}(user.Email, emailBody)
	}
}

func formatSkillsList(skills []string) string {
	if len(skills) == 0 {
		return ""
	}
	res := skills[0]
	for i := 1; i < len(skills); i++ {
		res += ", " + skills[i]
	}
	return res
}
