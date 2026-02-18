package expense

import (
	"context"
	"database/sql"
	"time"

	"search-job/pkg/postgres" // ваш пакет с БД
)

type Repository interface {
	Create(ctx context.Context, amount float64, currency string, occurredAt time.Time, comment *string) error
	GetAll(ctx context.Context) ([]Expense, error)
	GetByID(ctx context.Context, id int64) (*Expense, error)
	Update(ctx context.Context, id int64, amount *float64, currency *string, occurredAt *time.Time, comment *string) error
	SoftDelete(ctx context.Context, id int64) error
}

type Repo struct {
	db *postgres.DB
}

func NewRepo(db *postgres.DB) *Repo { // ожидает *postgres.DB
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, amount float64, currency string, occurredAt time.Time, comment *string) error {
	query := `
		INSERT INTO expenses (
			amount,
			currency,
			occurred_at,
			comment,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`

	_, err := r.db.Exec(ctx, query, amount, currency, occurredAt, comment)
	return err
}

func (r *Repo) GetAll(ctx context.Context) ([]Expense, error) {
	query := `
		SELECT
			id,
			amount,
			currency,
			occurred_at,
			comment,
			created_at,
			updated_at
		FROM expenses
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var e Expense
		err := rows.Scan(
			&e.ID,
			&e.Amount,
			&e.Currency,
			&e.OccurredAt,
			&e.Comment,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*Expense, error) {
	query := `
		SELECT
			id,
			amount,
			currency,
			occurred_at,
			comment,
			created_at,
			updated_at
		FROM expenses
		WHERE id = $1
	`

	var e Expense
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.Amount,
		&e.Currency,
		&e.OccurredAt,
		&e.Comment,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (r *Repo) Update(ctx context.Context, id int64, amount *float64, currency *string, occurredAt *time.Time, comment *string) error {
	query := `
		UPDATE expenses 
		SET 
			amount = COALESCE($1, amount),
			currency = COALESCE($2, currency),
			occurred_at = COALESCE($3, occurred_at),
			comment = COALESCE($4, comment),
			updated_at = NOW()
		WHERE id = $5
	`

	_, err := r.db.Exec(ctx, query, amount, currency, occurredAt, comment, id)
	return err
}

func (r *Repo) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE expenses 
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	// Проверяем, что запись была найдена и обновлена
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
