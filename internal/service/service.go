package service

import (
	"net/http"
	"strconv"
	"time"

	"search-job/internal/expense"
	"search-job/pkg/postgres"

	"github.com/labstack/echo/v4"
)

const (
	internalServerError = "internal error"
)

type Service struct {
	expenseRepo *expense.Repo
	logger      echo.Logger
}

func NewService(db *postgres.DB, logger echo.Logger) *Service {
	return &Service{
		expenseRepo: expense.NewRepo(db),
		logger:      logger,
	}
}

type Response struct {
	Object       any    `json:"object,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

// CreateExpense обрабатывает создание нового расхода
func (s *Service) CreateExpense(c echo.Context) error {
	type request struct {
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"` // добавили валюту
		OccurredAt string  `json:"occurred_at"`
		Comment    string  `json:"comment,omitempty"`
	}

	var req request
	if err := c.Bind(&req); err != nil {
		s.logger.Errorf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid params"})
	}

	// Валидация
	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "amount must be > 0"})
	}

	if len(req.Currency) != 3 {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "currency must be 3 letters"})
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		s.logger.Errorf("Failed to parse occurred_at: %v", err)
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid occurred_at format, use RFC3339"})
	}

	var commentPtr *string
	if req.Comment != "" {
		commentPtr = &req.Comment
	}

	err = s.expenseRepo.Create(
		c.Request().Context(),
		req.Amount,
		req.Currency,
		occurredAt,
		commentPtr,
	)
	if err != nil {
		s.logger.Errorf("Failed to create expense: %v", err)
		return c.JSON(http.StatusInternalServerError, Response{ErrorMessage: internalServerError})
	}

	return c.JSON(http.StatusCreated, Response{Object: map[string]string{"status": "created"}})
}

// GetExpenses возвращает список всех расходов
func (s *Service) GetExpenses(c echo.Context) error {
	expenses, err := s.expenseRepo.GetAll(c.Request().Context())
	if err != nil {
		s.logger.Errorf("Failed to get expenses: %v", err)
		return c.JSON(http.StatusInternalServerError, Response{ErrorMessage: internalServerError})
	}

	return c.JSON(http.StatusOK, Response{Object: expenses})
}

// GetExpenseByID возвращает расход по ID
func (s *Service) GetExpenseByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid id format"})
	}

	expense, err := s.expenseRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		s.logger.Errorf("Failed to get expense by ID %d: %v", id, err)
		return c.JSON(http.StatusNotFound, Response{ErrorMessage: "expense not found"})
	}

	return c.JSON(http.StatusOK, Response{Object: expense})
}

// UpdateExpense обновляет существующий расход
func (s *Service) UpdateExpense(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid id format"})
	}

	type request struct {
		Amount     *float64 `json:"amount,omitempty"`
		Currency   *string  `json:"currency,omitempty"`
		OccurredAt *string  `json:"occurred_at,omitempty"`
		Comment    *string  `json:"comment,omitempty"`
	}

	var req request
	if err := c.Bind(&req); err != nil {
		s.logger.Errorf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid params"})
	}

	// Валидация
	if req.Amount != nil && *req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "amount must be > 0"})
	}

	if req.Currency != nil && len(*req.Currency) != 3 {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "currency must be 3 letters"})
	}

	var occurredAt *time.Time
	if req.OccurredAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.OccurredAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid occurred_at format"})
		}
		occurredAt = &parsed
	}

	err = s.expenseRepo.Update(
		c.Request().Context(),
		id,
		req.Amount,
		req.Currency,
		occurredAt,
		req.Comment,
	)
	if err != nil {
		s.logger.Errorf("Failed to update expense %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, Response{ErrorMessage: internalServerError})
	}

	return c.JSON(http.StatusOK, Response{Object: map[string]string{"status": "updated"}})
}

// DeleteExpense выполняет мягкое удаление расхода
func (s *Service) DeleteExpense(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{ErrorMessage: "invalid id format"})
	}

	if err := s.expenseRepo.SoftDelete(c.Request().Context(), id); err != nil {
		s.logger.Errorf("Failed to delete expense %d: %v", id, err)
		return c.JSON(http.StatusInternalServerError, Response{ErrorMessage: internalServerError})
	}

	return c.JSON(http.StatusOK, Response{Object: map[string]string{"status": "deleted"}})
}
