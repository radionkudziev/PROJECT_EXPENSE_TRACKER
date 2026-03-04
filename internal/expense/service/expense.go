package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"search-job/internal/expense"
	"search-job/internal/middleware"
	"search-job/internal/models"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Service) CreateExpense(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		CategoryID *int64  `json:"category_id"`
		OccurredAt string  `json:"occurred_at"`
		Comment    string  `json:"comment"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	// Конвертация
	amountBase, err := s.convertAmount(c.Request().Context(), userID, req.Amount, req.Currency)
	if err != nil {
		s.logger.Errorf("conversion failed: %v", err)
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "currency conversion service unavailable"})
	}

	expense := &models.Expense{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		AmountBase: amountBase,
		Currency:   req.Currency,
		OccurredAt: occurredAt,
	}

	if req.Comment != "" {
		expense.Comment = &req.Comment
	}

	if err := s.expenseRepo.Create(c.Request().Context(), expense); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusCreated, expense)
}

func (s *Service) GetExpenses(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	params := expense.GetExpensesParams{
		UserID: userID,
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	params.Limit = limit
	params.Offset = (page - 1) * limit

	if from := c.QueryParam("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			params.From = &t
		}
	}
	if to := c.QueryParam("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			params.To = &t
		}
	}
	if catID := c.QueryParam("category_id"); catID != "" {
		id, _ := strconv.ParseInt(catID, 10, 64)
		params.CategoryID = &id
	}
	if min := c.QueryParam("min"); min != "" {
		val, _ := strconv.ParseFloat(min, 64)
		params.MinAmount = &val
	}
	if max := c.QueryParam("max"); max != "" {
		val, _ := strconv.ParseFloat(max, 64)
		params.MaxAmount = &val
	}
	params.Search = c.QueryParam("search")
	params.Sort = c.QueryParam("sort")
	params.Order = c.QueryParam("order")

	expenses, total, err := s.expenseRepo.GetAll(c.Request().Context(), params)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": expenses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (s *Service) GetExpenseByID(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	expense, err := s.expenseRepo.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, expense)
}

func (s *Service) UpdateExpense(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	var req struct {
		CategoryID *int64   `json:"category_id"`
		Amount     *float64 `json:"amount"`
		Currency   *string  `json:"currency"`
		OccurredAt *string  `json:"occurred_at"`
		Comment    *string  `json:"comment"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	var occurredAt *time.Time
	if req.OccurredAt != nil {
		t, err := time.Parse(time.RFC3339, *req.OccurredAt)
		if err != nil {
			return c.JSON(s.NewError(InvalidParams))
		}
		occurredAt = &t
	}

	expense := &models.Expense{
		ID:         id,
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount: func() float64 {
			if req.Amount != nil {
				return *req.Amount
			}
			return 0
		}(),
		Currency: func() string {
			if req.Currency != nil {
				return *req.Currency
			}
			return ""
		}(),
		OccurredAt: func() time.Time {
			if occurredAt != nil {
				return *occurredAt
			}
			return time.Time{}
		}(),
		Comment: req.Comment,
	}

	if err := s.expenseRepo.Update(c.Request().Context(), expense); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, expense)
}

func (s *Service) DeleteExpense(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	if err := s.expenseRepo.Delete(c.Request().Context(), id, userID); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

func (s *Service) convertAmount(ctx context.Context, userID int64, amount float64, fromCurrency string) (*float64, error) {
	// Получаем базовую валюту пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if fromCurrency == user.BaseCurrency {
		return &amount, nil
	}
	// Вызываем внешний API
	rate, err := s.exchangeClient.GetRate(fromCurrency, user.BaseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}
	converted := amount * rate
	return &converted, nil
}

func (s *Service) RestoreExpense(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(s.NewError(InvalidParams))
	}

	if err := s.expenseRepo.Restore(c.Request().Context(), id, userID); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "expense not found or not deleted"})
		}
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	// Получаем восстановленный расход
	expense, err := s.expenseRepo.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, expense)
}

func (s *Service) GetSummaryByCategories(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var from, to *time.Time
	if fromStr := c.QueryParam("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return c.JSON(s.NewError(InvalidParams))
		}
		from = &t
	}
	if toStr := c.QueryParam("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return c.JSON(s.NewError(InvalidParams))
		}
		to = &t
	}

	summaries, err := s.expenseRepo.GetSummaryByCategories(c.Request().Context(), userID, from, to)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, summaries)
}
