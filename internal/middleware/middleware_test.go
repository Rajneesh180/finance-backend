package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/google/uuid"
)

func TestAuth_MissingHeader(t *testing.T) {
	authSvc := service.NewAuthService("test-secret", 60)
	handler := Auth(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	authSvc := service.NewAuthService("test-secret", 60)
	handler := Auth(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	authSvc := service.NewAuthService("test-secret", 60)
	handler := Auth(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	authSvc := service.NewAuthService("test-secret", 60)
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  domain.RoleAnalyst,
	}
	token, _ := authSvc.GenerateToken(user)

	var gotClaims *service.Claims
	handler := Auth(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusOK)
	}
	if gotClaims == nil {
		t.Fatal("claims should not be nil")
	}
	if gotClaims.UserID != user.ID {
		t.Errorf("got user_id %v, want %v", gotClaims.UserID, user.ID)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	authSvc := service.NewAuthService("test-secret", 60)
	user := &domain.User{ID: uuid.New(), Email: "a@b.com", Role: domain.RoleAdmin}
	token, _ := authSvc.GenerateToken(user)

	inner := RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler := Auth(authSvc)(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	authSvc := service.NewAuthService("test-secret", 60)
	user := &domain.User{ID: uuid.New(), Email: "v@b.com", Role: domain.RoleViewer}
	token, _ := authSvc.GenerateToken(user)

	inner := RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))
	handler := Auth(authSvc)(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusForbidden)
	}
}
