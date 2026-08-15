package models

import "time"

// TransactionFilter holds all possible query parameters for history
type TransactionFilter struct {
	Page     int
	Limit    int
	Category string
	From     string
	To       string
}

// CategorySummary represents the aggregated totals grouped by category
type CategorySummary struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

// Budget represents a monthly limit for a specific category
type Budget struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id"`
	Category     string    `json:"category"`
	MonthlyLimit int64     `json:"monthly_limit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BudgetStatus is the response structure for the status endpoint
type BudgetStatus struct {
	Category     string `json:"category"`
	MonthlyLimit int64  `json:"monthly_limit"`
	SpentSoFar   int64  `json:"spent_so_far"`
	OverBudget   bool   `json:"over_budget"`
}
