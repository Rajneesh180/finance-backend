package service

import (
	"context"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/testutil"
)

func newTestUserService() (*UserService, *testutil.MockUserRepo) {
	repo := testutil.NewMockUserRepo()
	auth := NewAuthService("test-secret", 60)
	return NewUserService(repo, auth), repo
}

func TestRegister_Success(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	user, err := svc.Register(ctx, domain.CreateUserRequest{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("got email %q, want %q", user.Email, "new@example.com")
	}
	if user.Role != domain.RoleViewer {
		t.Errorf("got role %q, want %q", user.Role, domain.RoleViewer)
	}
	if !user.IsActive {
		t.Error("new user should be active")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, domain.CreateUserRequest{
		Email: "dup@example.com", Password: "password123", Name: "First",
	})

	_, err := svc.Register(ctx, domain.CreateUserRequest{
		Email: "dup@example.com", Password: "password456", Name: "Second",
	})
	if err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, domain.CreateUserRequest{
		Email: "login@example.com", Password: "password123", Name: "Login User",
	})

	token, user, err := svc.Login(ctx, domain.LoginRequest{
		Email: "login@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if user.Email != "login@example.com" {
		t.Errorf("got email %q, want %q", user.Email, "login@example.com")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, domain.CreateUserRequest{
		Email: "wrong@example.com", Password: "password123", Name: "User",
	})

	_, _, err := svc.Login(ctx, domain.LoginRequest{
		Email: "wrong@example.com", Password: "wrongpass",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_NonExistent(t *testing.T) {
	svc, _ := newTestUserService()
	_, _, err := svc.Login(context.Background(), domain.LoginRequest{
		Email: "nobody@example.com", Password: "password123",
	})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	user, _ := svc.Register(ctx, domain.CreateUserRequest{
		Email: "update@example.com", Password: "password123", Name: "Old Name",
	})

	updated, err := svc.UpdateProfile(ctx, user.ID, domain.UpdateProfileRequest{
		Name: "New Name",
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("got name %q, want %q", updated.Name, "New Name")
	}
}

func TestAdminUpdate_ChangeRole(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	user, _ := svc.Register(ctx, domain.CreateUserRequest{
		Email: "promote@example.com", Password: "password123", Name: "Promo User",
	})

	newRole := domain.RoleAnalyst
	updated, err := svc.AdminUpdate(ctx, user.ID, domain.AdminUpdateUserRequest{
		Role: &newRole,
	})
	if err != nil {
		t.Fatalf("AdminUpdate: %v", err)
	}
	if updated.Role != domain.RoleAnalyst {
		t.Errorf("got role %q, want %q", updated.Role, domain.RoleAnalyst)
	}
}

func TestDelete_Success(t *testing.T) {
	svc, _ := newTestUserService()
	ctx := context.Background()

	user, _ := svc.Register(ctx, domain.CreateUserRequest{
		Email: "delete@example.com", Password: "password123", Name: "Delete Me",
	})

	if err := svc.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := svc.GetByID(ctx, user.ID)
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound after delete, got %v", err)
	}
}
