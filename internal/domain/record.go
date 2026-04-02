package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type RecordType string

const (
	RecordIncome  RecordType = "income"
	RecordExpense RecordType = "expense"
)

func (t RecordType) IsValid() bool {
	return t == RecordIncome || t == RecordExpense
}

type FinancialRecord struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Type        RecordType      `json:"type"`
	Category    string          `json:"category"`
	Date        time.Time       `json:"date"`
	Description *string         `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"-"`
}

type CreateRecordRequest struct {
	Amount      string `json:"amount" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=income expense"`
	Category    string `json:"category" validate:"required,max=100"`
	Date        string `json:"date" validate:"required"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateRecordRequest struct {
	Amount      *string `json:"amount"`
	Type        *string `json:"type" validate:"omitempty,oneof=income expense"`
	Category    *string `json:"category" validate:"omitempty,max=100"`
	Date        *string `json:"date"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type RecordFilter struct {
	Type       string
	Category   string
	DateFrom   string
	DateTo     string
	Page       int
	PerPage    int
	SortBy     string
	SortOrder  string
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Dashboard aggregation types

type DashboardSummary struct {
	TotalIncome  decimal.Decimal  `json:"total_income"`
	TotalExpense decimal.Decimal  `json:"total_expense"`
	NetBalance   decimal.Decimal  `json:"net_balance"`
	RecordCount  int              `json:"record_count"`
	ByCategory   []CategoryTotal  `json:"by_category"`
	Trend        []TrendPoint     `json:"trend"`
}

type CategoryTotal struct {
	Category string          `json:"category"`
	Total    decimal.Decimal `json:"total"`
	Count    int             `json:"count"`
}

type TrendPoint struct {
	Period  string          `json:"period"`
	Income  decimal.Decimal `json:"income"`
	Expense decimal.Decimal `json:"expense"`
	Net     decimal.Decimal `json:"net"`
}
