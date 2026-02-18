package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"search-job/internal/config"
	"search-job/internal/expense"
	"search-job/internal/service"
	"search-job/pkg/logs"
	"search-job/pkg/postgres"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err) // временно, пока нет логгера
	}

	// 2. Инициализация логгера (с двумя параметрами)
	logger := logs.NewLogger(!cfg.IsProd, "info") // Исправлено!

	// 3. Подключение к БД
	ctx := context.Background()
	db, err := postgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()
	logger.Info("Database connected successfully")

	// 4. Создаем репозиторий
	//expenseRepo := expense.NewRepo(db)  // теперь должно работать
	_ = expense.NewRepo(db)
	// 5. Создаем сервис
	svc := service.NewService(db, logger)

	// 6. Настройка HTTP сервера
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | ${status} | ${method} | ${uri} | ${latency_human}\n",
	}))

	// 7. Регистрация маршрутов
	api := e.Group("/api/v1")

	api.POST("/expenses", svc.CreateExpense)
	api.GET("/expenses", svc.GetExpenses)
	api.GET("/expenses/:id", svc.GetExpenseByID)
	api.PUT("/expenses/:id", svc.UpdateExpense)
	api.DELETE("/expenses/:id", svc.DeleteExpense)

	// 8. Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 9. Запуск сервера
	go func() {
		port := cfg.GetWebPort()
		if port == "" {
			port = ":8080"
		}

		logger.Infof("Server starting on port %s", port)
		if err := e.Start(port); err != nil && err.Error() != "http: Server closed" {
			logger.Fatal("Server error: ", err)
		}
	}()

	// 10. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server shutdown error: ", err)
	}

	logger.Info("Server stopped")
}
