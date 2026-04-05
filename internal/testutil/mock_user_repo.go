package testutil

import (
	"context"
	"errors"
	"sync"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type MockUserRepo struct {
	mu    sync.Mutex
	users map[uuid.UUID]*domain.User
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (m *MockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *MockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

func (m *MockUserRepo) List(_ context.Context, page, perPage int) ([]domain.User, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []domain.User
	for _, u := range m.users {
		all = append(all, *u)
	}
	total := len(all)
	start := (page - 1) * perPage
	if start >= total {
		return nil, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}
