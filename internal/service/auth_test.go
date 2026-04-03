package service

import (
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	svc := NewAuthService("test-secret", 15)

	hash, err := svc.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "mypassword" {
		t.Fatal("hash should not equal plaintext")
	}
	if !svc.CheckPassword(hash, "mypassword") {
		t.Fatal("CheckPassword should return true for correct password")
	}
	if svc.CheckPassword(hash, "wrongpassword") {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	svc := NewAuthService("test-secret-key", 60)

	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  domain.RoleAnalyst,
	}

	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("got user_id %v, want %v", claims.UserID, user.ID)
	}
	if claims.Email != user.Email {
		t.Errorf("got email %q, want %q", claims.Email, user.Email)
	}
	if claims.Role != user.Role {
		t.Errorf("got role %q, want %q", claims.Role, user.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := NewAuthService("test-secret", 15)

	_, err := svc.ValidateToken("garbage.token.here")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewAuthService("secret-one", 15)
	svc2 := NewAuthService("secret-two", 15)

	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  domain.RoleViewer,
	}

	token, _ := svc1.GenerateToken(user)
	_, err := svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error when validating with different secret")
	}
}
