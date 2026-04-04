package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrEmailTaken         = errors.New("email already in use")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountDeactivated = errors.New("account is deactivated")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRole        = errors.New("invalid role")
)

type UserService struct {
	repo repository.UserRepository
	auth *AuthService
}

func NewUserService(repo repository.UserRepository, auth *AuthService) *UserService {
	return &UserService{repo: repo, auth: auth}
}

func (s *UserService) Register(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := s.auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: hash,
		Name:     req.Name,
		Role:     domain.RoleViewer,
		IsActive: true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (string, *domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, ErrUserNotFound
	}
	if !user.IsActive {
		return "", nil, ErrAccountDeactivated
	}
	if !s.auth.CheckPassword(user.Password, req.Password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.auth.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("generating token: %w", err)
	}
	return token, user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, req domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Email != "" && req.Email != user.Email {
		existing, _ := s.repo.GetByEmail(ctx, req.Email)
		if existing != nil {
			return nil, ErrEmailTaken
		}
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}
	return user, nil
}

func (s *UserService) AdminUpdate(ctx context.Context, id uuid.UUID, req domain.AdminUpdateUserRequest) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Role != nil {
		if !req.Role.IsValid() {
			return nil, ErrInvalidRole
		}
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrUserNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *UserService) List(ctx context.Context, page, perPage int) ([]domain.User, int, error) {
	return s.repo.List(ctx, page, perPage)
}
