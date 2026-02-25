package main

import (
	"context"
	"search-job/internal/auth"
	"search-job/internal/config"
	"search-job/internal/expense/service"
	"search-job/internal/middleware"
	"search-job/internal/pkg/logs"
	"search-job/internal/pkg/postgres"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	logger := logs.NewLogger(false)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err)
	}

	db, err := postgres.ConnectPostgres(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal(err)
	}

	svc := service.NewService(db, logger)
	authHandler := auth.NewHandler(db)

	router := echo.New()

	// Публичные маршруты
	auth := router.Group("/api/v1/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// Защищенные маршруты
	api := router.Group("/api/v1", middleware.AuthMiddleware)

	// Categories
	api.POST("/categories", svc.CreateCategory)
	api.GET("/categories", svc.GetCategories)
	api.PATCH("/categories/:id", svc.UpdateCategory)
	api.DELETE("/categories/:id", svc.DeleteCategory)

	// Expenses
	api.POST("/expenses", svc.CreateExpense)
	api.GET("/expenses", svc.GetExpenses)
	api.GET("/expenses/:id", svc.GetExpenseByID)
	api.PATCH("/expenses/:id", svc.UpdateExpense)
	api.DELETE("/expenses/:id", svc.DeleteExpense)

	router.Logger.Fatal(router.Start(cfg.GetWebPort()))
}
