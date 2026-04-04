package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.New()
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true

	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, password, name, role, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Email, user.Password, user.Name, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, name, role, is_active, created_at, updated_at
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, name, role, is_active, created_at, updated_at
		 FROM users WHERE email = $1 AND deleted_at IS NULL`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return &u, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $1, name = $2, role = $3, is_active = $4, updated_at = $5
		 WHERE id = $6 AND deleted_at IS NULL`,
		user.Email, user.Name, user.Role, user.IsActive, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("soft deleting user: %w", err)
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, page, perPage int) ([]domain.User, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	offset := (page - 1) * perPage
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, name, role, is_active, created_at, updated_at
		 FROM users WHERE deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, perPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, u)
	}
	return users, total, nil
}
