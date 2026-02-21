package main

import (
	"context"
	"search-job/internal/config"
	"search-job/internal/service"
	"search-job/pkg/logs"
	"search-job/pkg/postgres"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	logger := logs.NewLogger(false)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err)
	}

	db, err := postgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal(err)
	}

	log.Info("Postgres successfully connected")

	svc := service.NewService(db, logger)

	router := echo.New()

	api := router.Group("/api/v1")
	api.GET("/expenses/:id", svc.GetExpenseByID)
	api.GET("/expenses", svc.GetExpenses)
	api.POST("/expenses", svc.CreateExpense)
	api.PUT("/expenses/:id", svc.UpdateExpense)
	api.DELETE("/expenses/:id", svc.DeleteExpense)

	router.Logger.Fatal(router.Start(":" + cfg.GetWebPort()))
}
