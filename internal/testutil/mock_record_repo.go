package testutil

import (
	"context"
	"sync"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MockRecordRepo struct {
	mu      sync.Mutex
	records map[uuid.UUID]*domain.FinancialRecord
}

func NewMockRecordRepo() *MockRecordRepo {
	return &MockRecordRepo{records: make(map[uuid.UUID]*domain.FinancialRecord)}
}

func (m *MockRecordRepo) Create(_ context.Context, r *domain.FinancialRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[r.ID] = r
	return nil
}

func (m *MockRecordRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.FinancialRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[id]
	if !ok {
		return nil, context.Canceled
	}
	return r, nil
}

func (m *MockRecordRepo) Update(_ context.Context, r *domain.FinancialRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[r.ID] = r
	return nil
}

func (m *MockRecordRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, id)
	return nil
}

func (m *MockRecordRepo) List(_ context.Context, userID *uuid.UUID, filter domain.RecordFilter) ([]domain.FinancialRecord, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.FinancialRecord
	for _, r := range m.records {
		if userID != nil && r.UserID != *userID {
			continue
		}
		result = append(result, *r)
	}
	return result, len(result), nil
}

type MockDashboardRepo struct{}

func NewMockDashboardRepo() *MockDashboardRepo {
	return &MockDashboardRepo{}
}

func (m *MockDashboardRepo) GetSummary(_ context.Context, _ *uuid.UUID, _, _ string) (*domain.DashboardSummary, error) {
	return &domain.DashboardSummary{
		TotalIncome:  decimal.NewFromInt(5000),
		TotalExpense: decimal.NewFromInt(2000),
		NetBalance:   decimal.NewFromInt(3000),
		RecordCount:  5,
	}, nil
}
