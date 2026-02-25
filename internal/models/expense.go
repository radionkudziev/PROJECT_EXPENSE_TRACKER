package models

import "time"

type Expense struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	CategoryID *int64    `json:"category_id,omitempty" db:"category_id"`
	Amount     float64   `json:"amount" db:"amount"`
	Currency   string    `json:"currency" db:"currency"`
	OccurredAt time.Time `json:"occurred_at" db:"occurred_at"`
	Comment    *string   `json:"comment,omitempty" db:"comment"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type Category struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	Name      string     `json:"name" db:"name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
