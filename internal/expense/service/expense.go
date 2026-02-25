package service

import (
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
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
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

	expense := &models.Expense{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
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
