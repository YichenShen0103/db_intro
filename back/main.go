package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"db_intro_backend/config"
	"db_intro_backend/db"
	"db_intro_backend/handlers"
	"db_intro_backend/middleware"
	"db_intro_backend/services"
	"db_intro_backend/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Init DB
	db.InitDB(cfg)
	defer db.DB.Close()

	// Init Services
	emailService := services.NewEmailService(cfg)
	excelService := services.NewExcelService()

	// Init Handlers
	projectHandler := handlers.NewProjectHandler(emailService, excelService)

	// Start Scheduler
	if utils.GetEnv("ENABLE_EMAIL_SCHEDULER", "true") == "true" {
		startEmailFetchScheduler(emailService)
	}

	r := gin.Default()

	r.Static("/uploads", "./uploads")

	// API routes
	api := r.Group("/api")
	{
		// Health check
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// Auth
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Departments
			protected.GET("/departments", handlers.GetDepartments)

			// Teachers
			protected.GET("/teachers", handlers.GetTeachers)
			protected.POST("/teachers", handlers.CreateTeacher)
			protected.PUT("/teachers/:id", handlers.UpdateTeacher)
			protected.DELETE("/teachers/:id", handlers.DeleteTeacher)

			// Projects
			protected.GET("/projects", projectHandler.GetProjects)
			protected.GET("/projects/:id", projectHandler.GetProject)
			protected.POST("/projects", projectHandler.CreateProject)
			protected.POST("/projects/:id/members", projectHandler.AddProjectMembers)
			protected.POST("/projects/:id/dispatch", projectHandler.DispatchProject)
			protected.GET("/projects/:id/tracking", projectHandler.GetProjectTracking)
			protected.POST("/projects/:id/remind", projectHandler.RemindTeachers)
			protected.POST("/projects/:id/fetch-emails", projectHandler.FetchProjectEmails)
			protected.POST("/projects/:id/aggregate", projectHandler.AggregateData)
			protected.GET("/projects/:id/download", projectHandler.DownloadAggregated)
		}
	}

	log.Printf("Server starting on port %s...\n", cfg.Port)
	r.Run(":" + cfg.Port)
}

func startEmailFetchScheduler(emailService *services.EmailService) {
	// Get interval from environment (in minutes, default 10 minutes)
	intervalStr := utils.GetEnv("EMAIL_FETCH_INTERVAL", "10")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 10
	}

	log.Printf("Starting email fetch scheduler with %d minute interval", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	go func() {
		for range ticker.C {
			log.Println("Starting scheduled email fetch")
			if err := emailService.ProcessIncomingEmails(); err != nil {
				log.Printf("Failed to process incoming emails: %v", err)
			}
			log.Println("Scheduled email fetch completed")
		}
	}()
}
