package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rakib/job-portal-api/internal/config"
	"github.com/rakib/job-portal-api/internal/database"
	"github.com/rakib/job-portal-api/internal/handler"
	"github.com/rakib/job-portal-api/internal/repository"
	"github.com/rakib/job-portal-api/internal/router"
	"github.com/rakib/job-portal-api/internal/service"
)

func main() {
	log.Println("Starting Job Portal REST API Backend...")

	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Connect to MongoDB Atlas
	db := database.ConnectDB()

	// 3. Migrate MongoDB indexes on startup
	database.MigrateIndexes(db)

	// 4. Initialize Repository Layer
	userRepo := repository.NewUserRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	jobRepo := repository.NewJobRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	savedRepo := repository.NewSavedJobRepository(db)

	// 5. Initialize Service Layer
	storageService := service.NewStorageService()
	authService := service.NewAuthService(userRepo, companyRepo, profileRepo, db)
	jobService := service.NewJobService(jobRepo, companyRepo)
	appService := service.NewApplicationService(appRepo, jobRepo, companyRepo, profileRepo, userRepo, storageService)
	profileService := service.NewProfileService(profileRepo, userRepo, storageService)
	companyService := service.NewCompanyService(companyRepo, jobRepo, userRepo, storageService)
	savedJobService := service.NewSavedJobService(savedRepo, jobRepo, jobService, companyRepo)
	dashboardService := service.NewDashboardService(appRepo, jobRepo, savedRepo, profileRepo, appService, jobService)

	// 6. Initialize Handler Layer
	authHandler := handler.NewAuthHandler(authService)
	jobHandler := handler.NewJobHandler(jobService)
	appHandler := handler.NewApplicationHandler(appService)
	profileHandler := handler.NewProfileHandler(profileService)
	companyHandler := handler.NewCompanyHandler(companyService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	healthHandler := handler.NewHealthHandler()

	savedJobH := handler.NewSavedJobHandler(savedJobService)
	router.InitSavedJobHandler(savedJobH)

	// 7. Setup Router and Engine
	engine := router.SetupRouter(
		authHandler,
		jobHandler,
		appHandler,
		profileHandler,
		companyHandler,
		dashboardHandler,
		healthHandler,
	)

	// 8. Start HTTP Server with Graceful Shutdown
	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: engine,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("REST API server running on port :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// Wait for interruption signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API server gracefully...")

	// Timeout context for shutdown draining
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	database.DisconnectDB()

	log.Println("REST API Server exited successfully.")
}
