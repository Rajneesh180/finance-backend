package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type dashboardRepo struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) DashboardRepository {
	return &dashboardRepo{pool: pool}
}

func (r *dashboardRepo) GetSummary(ctx context.Context, userID *uuid.UUID, dateFrom, dateTo string) (*domain.DashboardSummary, error) {
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIdx := 1

	if userID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *userID)
		argIdx++
	}
	if dateFrom != "" {
		where = append(where, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		where = append(where, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, dateTo)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")
	summary := &domain.DashboardSummary{}

	totalsQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0),
			COUNT(*)
		FROM financial_records WHERE %s`, whereClause)

	var totalIncome, totalExpense decimal.Decimal
	err := r.pool.QueryRow(ctx, totalsQuery, args...).Scan(&totalIncome, &totalExpense, &summary.RecordCount)
	if err != nil {
		return nil, fmt.Errorf("querying totals: %w", err)
	}
	summary.TotalIncome = totalIncome
	summary.TotalExpense = totalExpense
	summary.NetBalance = totalIncome.Sub(totalExpense)

	catQuery := fmt.Sprintf(`
		SELECT category, SUM(amount), COUNT(*)
		FROM financial_records WHERE %s
		GROUP BY category ORDER BY SUM(amount) DESC`, whereClause)

	catRows, err := r.pool.Query(ctx, catQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying categories: %w", err)
	}
	defer catRows.Close()

	for catRows.Next() {
		var ct domain.CategoryTotal
		var total decimal.Decimal
		if err := catRows.Scan(&ct.Category, &total, &ct.Count); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		ct.Total = total
		summary.ByCategory = append(summary.ByCategory, ct)
	}

	trendQuery := fmt.Sprintf(`
		SELECT
			TO_CHAR(date, 'YYYY-MM') AS period,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM financial_records WHERE %s
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY period`, whereClause)

	trendRows, err := r.pool.Query(ctx, trendQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying trend: %w", err)
	}
	defer trendRows.Close()

	for trendRows.Next() {
		var tp domain.TrendPoint
		var income, expense decimal.Decimal
		if err := trendRows.Scan(&tp.Period, &income, &expense); err != nil {
			return nil, fmt.Errorf("scanning trend: %w", err)
		}
		tp.Income = income
		tp.Expense = expense
		tp.Net = income.Sub(expense)
		summary.Trend = append(summary.Trend, tp)
	}

	return summary, nil
}
