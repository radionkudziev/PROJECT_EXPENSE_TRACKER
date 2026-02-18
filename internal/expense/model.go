package expense

import "time"

type Expense struct {
	ID         int64     `json:"id" db:"id"`
	Amount     float64   `json:"amount" db:"amount" validate:"required,gt=0"`
	Currency   string    `json:"currency" db:"currency" validate:"required,len=3"`
	OccurredAt time.Time `json:"occurred_at" db:"occurred_at" validate:"required"`
	Comment    *string   `json:"comment,omitempty" db:"comment"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	// DeletedAt убрали для Уровня 1
}

// Для создания нового расхода (без ID и временных меток)
type CreateExpenseRequest struct {
	Amount     float64 `json:"amount" validate:"required,gt=0"`
	Currency   string  `json:"currency" validate:"required,len=3"`
	OccurredAt string  `json:"occurred_at" validate:"required"` // формат: "2006-01-02T15:04:05Z"
	Comment    *string `json:"comment"`
}

// Для обновления расхода
type UpdateExpenseRequest struct {
	Amount     *float64 `json:"amount" validate:"omitempty,gt=0"`
	Currency   *string  `json:"currency" validate:"omitempty,len=3"`
	OccurredAt *string  `json:"occurred_at"`
	Comment    *string  `json:"comment"`
}

// Для ответа с расходом (может содержать дополнительную информацию)
type ExpenseResponse struct {
	Expense
}

// Для списка расходов
type ExpenseListResponse struct {
	Expenses []Expense `json:"expenses"`
	Total    int64     `json:"total"`
}
