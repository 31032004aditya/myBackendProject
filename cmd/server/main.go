package main

import (
	"log"

	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// Instantiate Repositories (In-Memory)
	userRepo := repository.NewUserRepository()
	recordRepo := repository.NewRecordRepository()

	// Instantiate Handlers
	authHandler := handler.NewAuthHandler(userRepo)
	userHandler := handler.NewUserHandler(userRepo)
	recordHandler := handler.NewRecordHandler(recordRepo)

	// Setup Gin router
	r := gin.Default()

	// Public Routes
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// Protected Routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		// Dashboard Endpoints (Analyst, Admin)
		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.RoleRequired("analyst", "admin"))
		{
			dashboard.GET("/summary", recordHandler.GetSummary)
			dashboard.GET("/category-totals", recordHandler.GetCategoryTotals)
		}

		// Record Management
		records := api.Group("/records")
		{
			// Analysts and Admins can view/list records
			records.GET("", middleware.RoleRequired("analyst", "admin"), recordHandler.List)
			records.GET("/:id", middleware.RoleRequired("analyst", "admin"), recordHandler.Get)

			// Only Admins can modify records
			records.POST("", middleware.RoleRequired("admin"), recordHandler.Create)
			records.PUT("/:id", middleware.RoleRequired("admin"), recordHandler.Update)
			records.DELETE("/:id", middleware.RoleRequired("admin"), recordHandler.Delete)
		}

		// User Management (Admin Only)
		users := api.Group("/users")
		users.Use(middleware.RoleRequired("admin"))
		{
			users.GET("", userHandler.ListUsers)
			users.PUT("/:id/role", userHandler.UpdateRole)
		}
	}

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
