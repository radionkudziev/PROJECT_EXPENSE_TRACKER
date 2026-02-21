package service

import (
	"net/http"
	"search-job/internal/models"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Service) CreateExpense(c echo.Context) error {
	var expense models.Expense

	err := c.Bind(&expense)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	repo := s.expenseRepo
	err = repo.Create(c.Request().Context(), &expense)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, Response{Object: expense})
}

func (s *Service) GetExpenses(c echo.Context) error {
	repo := s.expenseRepo

	expenses, err := repo.GetAll(c.Request().Context())
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, Response{Object: expenses})
}

func (s *Service) GetExpenseByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	repo := s.expenseRepo

	expense, err := repo.GetByID(c.Request().Context(), int64(id))
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, Response{Object: expense})
}

func (s *Service) UpdateExpense(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	var expense models.Expense
	err = c.Bind(&expense)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	expense.ID = int64(id)
	expense.UpdatedAt = time.Now()

	repo := s.expenseRepo
	err = repo.Update(c.Request().Context(), &expense)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, Response{Object: expense})
}

func (s *Service) DeleteExpense(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	repo := s.expenseRepo
	err = repo.SoftDelete(c.Request().Context(), int64(id))
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, Response{Object: "deleted"})
}
