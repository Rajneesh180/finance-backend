package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type recordRepo struct {
	pool *pgxpool.Pool
}

func NewRecordRepository(pool *pgxpool.Pool) RecordRepository {
	return &recordRepo{pool: pool}
}

func (r *recordRepo) Create(ctx context.Context, rec *domain.FinancialRecord) error {
	rec.ID = uuid.New()
	now := time.Now().UTC()
	rec.CreatedAt = now
	rec.UpdatedAt = now

	_, err := r.pool.Exec(ctx,
		`INSERT INTO financial_records (id, user_id, amount, type, category, date, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		rec.ID, rec.UserID, rec.Amount, rec.Type, rec.Category, rec.Date, rec.Description, rec.CreatedAt, rec.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting record: %w", err)
	}
	return nil
}

func (r *recordRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.FinancialRecord, error) {
	var rec domain.FinancialRecord
	var amount decimal.Decimal
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, amount, type, category, date, description, created_at, updated_at
		 FROM financial_records WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&rec.ID, &rec.UserID, &amount, &rec.Type, &rec.Category, &rec.Date, &rec.Description, &rec.CreatedAt, &rec.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying record by id: %w", err)
	}
	rec.Amount = amount
	return &rec, nil
}

func (r *recordRepo) Update(ctx context.Context, rec *domain.FinancialRecord) error {
	rec.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE financial_records
		 SET amount = $1, type = $2, category = $3, date = $4, description = $5, updated_at = $6
		 WHERE id = $7 AND deleted_at IS NULL`,
		rec.Amount, rec.Type, rec.Category, rec.Date, rec.Description, rec.UpdatedAt, rec.ID,
	)
	if err != nil {
		return fmt.Errorf("updating record: %w", err)
	}
	return nil
}

func (r *recordRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE financial_records SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("soft deleting record: %w", err)
	}
	return nil
}

func (r *recordRepo) List(ctx context.Context, userID *uuid.UUID, filter domain.RecordFilter) ([]domain.FinancialRecord, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIdx := 1

	if userID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *userID)
		argIdx++
	}
	if filter.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.DateFrom != "" {
		where = append(where, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		where = append(where, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT count(*) FROM financial_records WHERE %s", whereClause)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting records: %w", err)
	}

	sortCol := "created_at"
	switch filter.SortBy {
	case "amount", "date", "type", "category":
		sortCol = filter.SortBy
	}
	sortDir := "DESC"
	if filter.SortOrder == "asc" {
		sortDir = "ASC"
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, amount, type, category, date, description, created_at, updated_at
		 FROM financial_records WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		whereClause, sortCol, sortDir, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing records: %w", err)
	}
	defer rows.Close()

	var records []domain.FinancialRecord
	for rows.Next() {
		var rec domain.FinancialRecord
		var amount decimal.Decimal
		if err := rows.Scan(&rec.ID, &rec.UserID, &amount, &rec.Type, &rec.Category, &rec.Date, &rec.Description, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning record row: %w", err)
		}
		rec.Amount = amount
		records = append(records, rec)
	}
	return records, total, nil
}
