package category

import (
	"context"
	"database/sql"
	"search-job/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (user_id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query, category.UserID, category.Name).Scan(
		&category.ID, &category.CreatedAt, &category.UpdatedAt,
	)
}

func (r *Repo) GetAll(ctx context.Context, userID int64, limit, offset int, search string) ([]models.Category, int, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM categories 
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	if search != "" {
		countQuery += " AND name ILIKE '%' || $2 || '%'"
	}

	var total int
	var err error
	if search != "" {
		err = r.db.QueryRow(ctx, countQuery, userID, search).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	if search != "" {
		query += " AND name ILIKE '%' || $2 || '%'"
	}
	query += " ORDER BY name LIMIT $3 OFFSET $4"

	var rows pgx.Rows
	if search != "" {
		rows, err = r.db.Query(ctx, query, userID, search, limit, offset)
	} else {
		rows, err = r.db.Query(ctx, query, userID, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		categories = append(categories, c)
	}

	return categories, total, nil
}

func (r *Repo) GetByID(ctx context.Context, id, userID int64) (*models.Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var c models.Category
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *Repo) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, category.Name, category.ID, category.UserID)
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
		UPDATE categories
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
