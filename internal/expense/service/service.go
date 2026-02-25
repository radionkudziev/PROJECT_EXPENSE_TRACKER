package service

import (
	"search-job/internal/category"
	"search-job/internal/expense"
	"search-job/internal/user"

	"github.com/labstack/gommon/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	InvalidParams       = "invalid params"
	InternalServerError = "internal error"
)

type Service struct {
	db           *pgxpool.Pool
	logger       *log.Logger // используем *logs.Logger
	expenseRepo  *expense.Repo
	userRepo     *user.Repo
	categoryRepo *category.Repo
}

func NewService(db *pgxpool.Pool, logger *log.Logger) *Service {
	svc := &Service{
		db:     db,
		logger: logger,
	}
	svc.initRepositories()
	return svc
}

func (s *Service) initRepositories() {
	s.expenseRepo = expense.NewRepo(s.db)
	s.userRepo = user.NewRepo(s.db)
	s.categoryRepo = category.NewRepo(s.db)
}

type Response struct {
	Object       any    `json:"object,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

func (r *Response) Error() string {
	return r.ErrorMessage
}

func (s *Service) NewError(err string) (int, *Response) {
	return 400, &Response{ErrorMessage: err}
}
