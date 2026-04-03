package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrNotOwner       = errors.New("not the record owner")
	ErrInvalidAmount  = errors.New("invalid amount")
	ErrInvalidDate    = errors.New("invalid date format, use YYYY-MM-DD")
)

type RecordService struct {
	repo      repository.RecordRepository
	dashboard repository.DashboardRepository
}

func NewRecordService(repo repository.RecordRepository, dash repository.DashboardRepository) *RecordService {
	return &RecordService{repo: repo, dashboard: dash}
}

func (s *RecordService) Create(ctx context.Context, userID uuid.UUID, req domain.CreateRecordRequest) (*domain.FinancialRecord, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || !amount.IsPositive() {
		return nil, ErrInvalidAmount
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	record := &domain.FinancialRecord{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      amount,
		Type:        domain.RecordType(req.Type),
		Category:    req.Category,
		Date:        date,
		Description: desc,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("creating record: %w", err)
	}
	return record, nil
}

func (s *RecordService) GetByID(ctx context.Context, id, userID uuid.UUID, isAdmin bool) (*domain.FinancialRecord, error) {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	if !isAdmin && record.UserID != userID {
		return nil, ErrNotOwner
	}
	return record, nil
}

func (s *RecordService) Update(ctx context.Context, id, userID uuid.UUID, isAdmin bool, req domain.UpdateRecordRequest) (*domain.FinancialRecord, error) {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	if !isAdmin && record.UserID != userID {
		return nil, ErrNotOwner
	}

	if req.Amount != nil {
		amount, err := decimal.NewFromString(*req.Amount)
		if err != nil || !amount.IsPositive() {
			return nil, ErrInvalidAmount
		}
		record.Amount = amount
	}
	if req.Type != nil {
		rt := domain.RecordType(*req.Type)
		if !rt.IsValid() {
			return nil, errors.New("invalid record type")
		}
		record.Type = rt
	}
	if req.Category != nil {
		record.Category = *req.Category
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, ErrInvalidDate
		}
		record.Date = date
	}
	if req.Description != nil {
		record.Description = req.Description
	}

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("updating record: %w", err)
	}
	return record, nil
}

func (s *RecordService) Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrRecordNotFound
	}
	if !isAdmin && record.UserID != userID {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

func (s *RecordService) List(ctx context.Context, userID *uuid.UUID, filter domain.RecordFilter) ([]domain.FinancialRecord, int, error) {
	return s.repo.List(ctx, userID, filter)
}

func (s *RecordService) Dashboard(ctx context.Context, userID *uuid.UUID, dateFrom, dateTo string) (*domain.DashboardSummary, error) {
	return s.dashboard.GetSummary(ctx, userID, dateFrom, dateTo)
}
