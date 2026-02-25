package service

import (
	"net/http"
	"search-job/internal/middleware"
	"search-job/internal/models"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (s *Service) CreateCategory(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	var category models.Category
	if err := c.Bind(&category); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	category.UserID = userID

	if err := s.categoryRepo.Create(c.Request().Context(), &category); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusCreated, category)
}

func (s *Service) GetCategories(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	search := c.QueryParam("search")

	categories, total, err := s.categoryRepo.GetAll(
		c.Request().Context(),
		userID,
		limit,
		offset,
		search,
	)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": categories,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (s *Service) UpdateCategory(c echo.Context) error {
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

	var category models.Category
	if err := c.Bind(&category); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InvalidParams))
	}

	category.ID = id
	category.UserID = userID

	if err := s.categoryRepo.Update(c.Request().Context(), &category); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, category)
}

func (s *Service) DeleteCategory(c echo.Context) error {
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

	if err := s.categoryRepo.Delete(c.Request().Context(), id, userID); err != nil {
		s.logger.Error(err)
		return c.JSON(s.NewError(InternalServerError))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}
