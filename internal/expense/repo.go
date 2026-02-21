package expense

import (
	"context"
	"search-job/internal/models"
	"search-job/pkg/postgres"
)

type Repo struct {
	db *postgres.DB
}

func NewRepo(db *postgres.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, expense *models.Expense) error {
	query := `
		INSERT INTO expenses (amount, currency, occurred_at, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	return r.db.QueryRow(ctx, query,
		expense.Amount,
		expense.Currency,
		expense.OccurredAt,
		expense.Comment,
		expense.CreatedAt,
		expense.UpdatedAt,
	).Scan(&expense.ID)
}

func (r *Repo) GetAll(ctx context.Context) ([]models.Expense, error) {
	query := `
		SELECT id, amount, currency, occurred_at, comment, created_at, updated_at
		FROM expenses
		WHERE deleted_at IS NULL
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
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

	return expenses, rows.Err()
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	query := `
		SELECT id, amount, currency, occurred_at, comment, created_at, updated_at
		FROM expenses
		WHERE id = $1 AND deleted_at IS NULL
	`

	var e models.Expense
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

func (r *Repo) Update(ctx context.Context, expense *models.Expense) error {
	query := `
		UPDATE expenses
		SET amount = $1, currency = $2, occurred_at = $3, comment = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		expense.Amount,
		expense.Currency,
		expense.OccurredAt,
		expense.Comment,
		expense.UpdatedAt,
		expense.ID,
	)
	return err
}

func (r *Repo) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE expenses
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}
