package main

import (
	"context"
	"search-job/internal/auth"
	"search-job/internal/config"
	"search-job/internal/expense/service"
	"search-job/internal/middleware"
	"search-job/internal/pkg/exchangerate"
	"search-job/internal/pkg/jwt"
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

	// Инициализация JWT
	jwt.Init(cfg.JWT.Secret)

	db, err := postgres.ConnectPostgres(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal(err)
	}

	// Создаём клиент для курсов валют
	exchangeClient := exchangerate.NewClient(
		cfg.ExternalAPI.CurrencyURL,
		cfg.ExternalAPI.APIKey,
		cfg.ExternalAPI.Timeout,
	)

	svc := service.NewService(db, logger, exchangeClient)
	authHandler := auth.NewHandler(db)

	router := echo.New()

	// Публичные маршруты
	auth := router.Group("/api/v1/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// Защищенные маршруты
	api := router.Group("/api/v1", middleware.AuthMiddleware)

	api.POST("/categories", svc.CreateCategory)
	api.GET("/categories", svc.GetCategories)
	api.PATCH("/categories/:id", svc.UpdateCategory)
	api.DELETE("/categories/:id", svc.DeleteCategory)

	api.POST("/expenses", svc.CreateExpense)
	api.GET("/expenses", svc.GetExpenses)
	api.GET("/expenses/:id", svc.GetExpenseByID)
	api.PATCH("/expenses/:id", svc.UpdateExpense)
	api.DELETE("/expenses/:id", svc.DeleteExpense)
	api.POST("/expenses/:id/restore", svc.RestoreExpense) // новый эндпоинт

	api.GET("/expenses/summary/by-categories", svc.GetSummaryByCategories) // новый эндпоинт

	router.Logger.Fatal(router.Start(cfg.GetWebPort()))
}
