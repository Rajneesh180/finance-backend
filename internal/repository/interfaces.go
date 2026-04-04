package repository

import (
	"context"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, perPage int) ([]domain.User, int, error)
}

type RecordRepository interface {
	Create(ctx context.Context, record *domain.FinancialRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FinancialRecord, error)
	Update(ctx context.Context, record *domain.FinancialRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, filter domain.RecordFilter) ([]domain.FinancialRecord, int, error)
}

type DashboardRepository interface {
	GetSummary(ctx context.Context, userID *uuid.UUID, dateFrom, dateTo string) (*domain.DashboardSummary, error)
	RecentActivity(ctx context.Context, userID *uuid.UUID, limit int) ([]domain.ActivityEntry, error)
}
