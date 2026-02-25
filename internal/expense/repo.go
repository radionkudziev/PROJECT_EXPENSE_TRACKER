package expense

import (
	"context"
	"database/sql"
	"fmt"
	"search-job/internal/models"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type GetExpensesParams struct {
	UserID     int64
	From       *time.Time
	To         *time.Time
	CategoryID *int64
	MinAmount  *float64
	MaxAmount  *float64
	Search     string
	Sort       string
	Order      string
	Limit      int
	Offset     int
}

func (r *Repo) Create(ctx context.Context, expense *models.Expense) error {
	query := `
		INSERT INTO expenses (user_id, category_id, amount, currency, occurred_at, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		expense.UserID,
		expense.CategoryID,
		expense.Amount,
		expense.Currency,
		expense.OccurredAt,
		expense.Comment,
	).Scan(&expense.ID, &expense.CreatedAt, &expense.UpdatedAt)
}

func (r *Repo) GetAll(ctx context.Context, params GetExpensesParams) ([]models.Expense, int, error) {

	where := []string{"e.user_id = $1", "e.deleted_at IS NULL"}
	args := []interface{}{params.UserID}
	argPos := 2

	if params.From != nil {
		where = append(where, fmt.Sprintf("e.occurred_at >= $%d", argPos))
		args = append(args, *params.From)
		argPos++
	}
	if params.To != nil {
		where = append(where, fmt.Sprintf("e.occurred_at <= $%d", argPos))
		args = append(args, *params.To)
		argPos++
	}
	if params.CategoryID != nil {
		where = append(where, fmt.Sprintf("e.category_id = $%d", argPos))
		args = append(args, *params.CategoryID)
		argPos++
	}
	if params.MinAmount != nil {
		where = append(where, fmt.Sprintf("e.amount >= $%d", argPos))
		args = append(args, *params.MinAmount)
		argPos++
	}
	if params.MaxAmount != nil {
		where = append(where, fmt.Sprintf("e.amount <= $%d", argPos))
		args = append(args, *params.MaxAmount)
		argPos++
	}
	if params.Search != "" {
		where = append(where, fmt.Sprintf("e.comment ILIKE '%%' || $%d || '%%'", argPos))
		args = append(args, params.Search)
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM expenses e
		WHERE %s
	`, whereClause)

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortField := "e.occurred_at"
	if params.Sort == "amount" {
		sortField = "e.amount"
	}
	sortOrder := "DESC"
	if params.Order == "asc" {
		sortOrder = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.category_id, e.amount, e.currency, 
		       e.occurred_at, e.comment, e.created_at, e.updated_at,
		       c.name as category_name
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id AND c.deleted_at IS NULL
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argPos, argPos+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		var categoryName *string
		err := rows.Scan(
			&e.ID, &e.UserID, &e.CategoryID, &e.Amount, &e.Currency,
			&e.OccurredAt, &e.Comment, &e.CreatedAt, &e.UpdatedAt,
			&categoryName,
		)
		if err != nil {
			return nil, 0, err
		}
		expenses = append(expenses, e)
	}

	return expenses, total, nil
}

func (r *Repo) GetByID(ctx context.Context, id, userID int64) (*models.Expense, error) {
	query := `
		SELECT e.id, e.user_id, e.category_id, e.amount, e.currency, 
		       e.occurred_at, e.comment, e.created_at, e.updated_at,
		       c.name as category_name
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id AND c.deleted_at IS NULL
		WHERE e.id = $1 AND e.user_id = $2 AND e.deleted_at IS NULL
	`

	var e models.Expense
	var categoryName *string
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&e.ID, &e.UserID, &e.CategoryID, &e.Amount, &e.Currency,
		&e.OccurredAt, &e.Comment, &e.CreatedAt, &e.UpdatedAt,
		&categoryName,
	)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *Repo) Update(ctx context.Context, expense *models.Expense) error {
	query := `
		UPDATE expenses
		SET category_id = COALESCE($1, category_id),
		    amount = COALESCE($2, amount),
		    currency = COALESCE($3, currency),
		    occurred_at = COALESCE($4, occurred_at),
		    comment = COALESCE($5, comment),
		    updated_at = NOW()
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		expense.CategoryID,
		expense.Amount,
		expense.Currency,
		expense.OccurredAt,
		expense.Comment,
		expense.ID,
		expense.UserID,
	)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, id, userID int64) error {
	query := `
		UPDATE expenses
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
