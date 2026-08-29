package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rakib/job-portal-api/internal/config"
	"github.com/rakib/job-portal-api/internal/handler"
	"github.com/rakib/job-portal-api/internal/middleware"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	jobHandler *handler.JobHandler,
	appHandler *handler.ApplicationHandler,
	profileHandler *handler.ProfileHandler,
	companyHandler *handler.CompanyHandler,
	dashboardHandler *handler.DashboardHandler,
	healthHandler *handler.HealthHandler,
	chatHandler *handler.ChatHandler,
	notifHandler *handler.NotificationHandler,
) *gin.Engine {
	r := gin.New()

	// Apply global middlewares
	r.Use(middleware.Recover())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Static serving of local uploads
	r.Static("/uploads", config.AppConfig.UploadDir)

	// Health check endpoints (Public)
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)

	// API versioning group
	api := r.Group("/api/v1")
	{
		// 1. Auth routes (Public / Rate limited)
		auth := api.Group("/auth")
		{
			auth.POST("/register/seeker", middleware.RateLimit(), authHandler.RegisterSeeker)
			auth.POST("/register/employer", middleware.RateLimit(), authHandler.RegisterEmployer)
			auth.POST("/login", middleware.RateLimit(), authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", middleware.AuthRequired(), authHandler.Logout)
			auth.GET("/me", middleware.AuthRequired(), authHandler.Me)
			auth.POST("/forgot-password", middleware.RateLimit(), authHandler.ForgotPassword)
			auth.POST("/reset-password", middleware.RateLimit(), authHandler.ResetPassword)
		}

		// 2. Public Jobs routes (Public)
		jobs := api.Group("/jobs")
		{
			jobs.GET("", jobHandler.ListPublicJobs)
			jobs.GET("/:id", jobHandler.GetJobByID)
			jobs.GET("/:id/similar", jobHandler.GetSimilarJobs)
			jobs.POST("/:id/report", middleware.OptionalAuth(), jobHandler.ReportJob)

			// Application Apply (Job Seeker role required / Rate limited)
			jobs.POST("/:id/applications", middleware.AuthRequired(), middleware.RequireRole("user"), middleware.RateLimit(), appHandler.ApplyToJob)
		}

		// 3. Seeker Applications routes (Job Seeker role required)
		apps := api.Group("/applications")
		apps.Use(middleware.AuthRequired(), middleware.RequireRole("user"))
		{
			apps.GET("/me", appHandler.ListSeekerApplications)
			apps.GET("/:id", appHandler.GetSeekerApplicationByID)
			apps.POST("/:id/withdraw", appHandler.WithdrawApplication)
		}

		// 4. Saved Jobs routes (Job Seeker role required)
		saved := api.Group("/saved-jobs")
		saved.Use(middleware.AuthRequired(), middleware.RequireRole("user"))
		{
			saved.GET("", savedJobHandler.GetSavedJobs)
			saved.POST("", savedJobHandler.SaveJob)
			saved.DELETE("/:jobId", savedJobHandler.UnsaveJob)
		}

		// 5. Seeker Profile routes (Job Seeker role required)
		prof := api.Group("/profile/me")
		prof.Use(middleware.AuthRequired(), middleware.RequireRole("user"))
		{
			prof.GET("", profileHandler.GetProfileMe)
			prof.PUT("", profileHandler.UpdateProfileMe)
			prof.POST("/avatar", profileHandler.UploadAvatar)
			prof.DELETE("/avatar", profileHandler.RemoveAvatar)
			prof.POST("/resume", profileHandler.UploadResume)
			prof.DELETE("/resume", profileHandler.RemoveResume)

			// Seeker Experience
			prof.POST("/experience", profileHandler.AddExperience)
			prof.PUT("/experience/:expId", profileHandler.UpdateExperience)
			prof.DELETE("/experience/:expId", profileHandler.DeleteExperience)

			// Seeker Education
			prof.POST("/education", profileHandler.AddEducation)
			prof.PUT("/education/:eduId", profileHandler.UpdateEducation)
			prof.DELETE("/education/:eduId", profileHandler.DeleteEducation)
		}

		// 6. Dashboards
		dash := api.Group("/dashboard")
		dash.Use(middleware.AuthRequired())
		{
			dash.GET("/seeker", middleware.RequireRole("user"), dashboardHandler.GetSeekerDashboard)
			dash.GET("/company", middleware.RequireRole("company"), dashboardHandler.GetCompanyDashboard)
		}

		// 7. Company Manage Jobs (Employer role required)
		cJobs := api.Group("/company/jobs")
		cJobs.Use(middleware.AuthRequired(), middleware.RequireRole("company"))
		{
			cJobs.GET("", jobHandler.ListCompanyJobs)
			cJobs.POST("", jobHandler.CreateJob)
			cJobs.GET("/:id", jobHandler.GetJobByID)
			cJobs.PUT("/:id", jobHandler.UpdateJob)
			cJobs.DELETE("/:id", jobHandler.DeleteJob)
			cJobs.POST("/:id/publish", jobHandler.PublishJob)
			cJobs.POST("/:id/close", jobHandler.CloseJob)
			cJobs.POST("/:id/reactivate", jobHandler.ReactivateJob)
			cJobs.POST("/bulk", jobHandler.BulkAction)
		}

		// 8. Company Applicants management (Employer role required)
		cApps := api.Group("/company/applicants")
		cApps.Use(middleware.AuthRequired(), middleware.RequireRole("company"))
		{
			cApps.GET("", appHandler.ListCompanyApplicants)
			cApps.GET("/:id", appHandler.GetCompanyApplicantByID)
			cApps.PATCH("/:id/status", appHandler.UpdateApplicantStatus)
			cApps.GET("/:id/resume", appHandler.DownloadResume)
		}

		// 9. Company settings & logo (Employer role required)
		cSettings := api.Group("/company")
		{
			cSettings.GET("/profile", middleware.AuthRequired(), middleware.RequireRole("company"), companyHandler.GetOwnCompanyProfile)
			cSettings.PUT("/profile", middleware.AuthRequired(), middleware.RequireRole("company"), companyHandler.UpdateCompanySettings)
			cSettings.GET("/settings", middleware.AuthRequired(), middleware.RequireRole("company"), companyHandler.GetCompanySettings)
			cSettings.PUT("/settings", middleware.AuthRequired(), middleware.RequireRole("company"), companyHandler.UpdateCompanySettings)
			cSettings.POST("/logo", middleware.AuthRequired(), middleware.RequireRole("company"), companyHandler.UploadLogo)
			cSettings.DELETE("/logo", middleware.AuthRequired(), middleware.RequireRole("company"), companyHandler.RemoveLogo)
		}

		// 10. Public companies profile (Public)
		api.GET("/companies", companyHandler.ListCompanies)
		api.GET("/companies/:id", companyHandler.GetPublicCompanyProfile)

		// 11. Chat routes (Authenticated)
		chats := api.Group("/chats")
		chats.Use(middleware.AuthRequired())
		{
			chats.POST("", chatHandler.CreateOrOpenRoom)
			chats.GET("", chatHandler.ListUserRooms)
			chats.GET("/:roomId/messages", chatHandler.GetRoomMessages)
		}

		// 12. WebSocket endpoint (Public routing, internally authenticated)
		api.GET("/chats/ws", chatHandler.HandleGlobalWebSocket)
		api.GET("/chats/:roomId/ws", chatHandler.HandleWebSocket)

		// 13. Notification routes (Authenticated)
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthRequired())
		{
			notifications.GET("", notifHandler.GetMyNotifications)
			notifications.PATCH("/:id/read", notifHandler.MarkAsRead)
		}
	}

	// Own job detail for edit is mounted inside cJobs group

	return r
}

var savedJobHandler *handler.SavedJobHandler

func InitSavedJobHandler(h *handler.SavedJobHandler) {
	savedJobHandler = h
}
