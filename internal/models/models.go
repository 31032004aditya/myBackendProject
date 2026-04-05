package models

import (
	"time"
)

// User represents the user in the system
type User struct {
	ID        uint           `json:"id"`
	Username  string         `json:"username"`
	Password  string         `json:"-"`
	Role      string         `json:"role"` // viewer, analyst, admin
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Record represents a financial entry
type Record struct {
	ID        uint           `json:"id"`
	Amount    float64        `json:"amount"`
	Type      string         `json:"type"` // income, expense
	Category  string         `json:"category"`
	Date      time.Time      `json:"date"`
	Notes     string         `json:"notes"`
	UserID    uint           `json:"userId"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// SummaryData represents the aggregated dashboard data
type SummaryData struct {
	TotalIncome  float64 `json:"totalIncome"`
	TotalExpense float64 `json:"totalExpense"`
	NetBalance   float64 `json:"netBalance"`
}

// CategoryTotal represents total per category
type CategoryTotal struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}
