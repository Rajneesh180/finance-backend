package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/Rajneesh180/finance-backend/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newTestUserHandler() (*UserHandler, *service.UserService, *service.AuthService) {
	repo := testutil.NewMockUserRepo()
	auth := service.NewAuthService("user-handler-test-secret", 60)
	userSvc := service.NewUserService(repo, auth)
	return NewUserHandler(userSvc), userSvc, auth
}

func registerUser(t *testing.T, userSvc *service.UserService, email, name string) *domain.User {
	t.Helper()
	user, err := userSvc.Register(context.Background(), domain.CreateUserRequest{
		Email:    email,
		Password: "password123",
		Name:     name,
	})
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}
	return user
}

func TestUserHandler_GetProfile(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	user := registerUser(t, userSvc, "profile@test.com", "Profile User")

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = withClaims(req, user.ID, domain.RoleViewer)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["email"] != "profile@test.com" {
		t.Errorf("got email %v, want profile@test.com", data["email"])
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	user := registerUser(t, userSvc, "update@test.com", "Before Update")

	body := `{"name":"After Update"}`
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, user.ID, domain.RoleViewer)
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "After Update" {
		t.Errorf("got name %v, want After Update", data["name"])
	}
}

func TestUserHandler_UpdateProfile_DuplicateEmail(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	registerUser(t, userSvc, "existing@test.com", "Existing")
	user := registerUser(t, userSvc, "updater@test.com", "Updater")

	body := `{"email":"existing@test.com"}`
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, user.ID, domain.RoleViewer)
	rr := httptest.NewRecorder()

	h.UpdateProfile(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusConflict)
	}
}

func TestUserHandler_AdminUpdateUser(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	admin := registerUser(t, userSvc, "admin@test.com", "Admin")
	target := registerUser(t, userSvc, "target@test.com", "Target")

	body := `{"role":"analyst"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+target.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, admin.ID, domain.RoleAdmin)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", target.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.AdminUpdateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["role"] != "analyst" {
		t.Errorf("got role %v, want analyst", data["role"])
	}
}

func TestUserHandler_AdminUpdate_NotFound(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	admin := registerUser(t, userSvc, "admin2@test.com", "Admin")

	fakeID := uuid.New()
	body := `{"role":"analyst"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+fakeID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, admin.ID, domain.RoleAdmin)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fakeID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.AdminUpdateUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	admin := registerUser(t, userSvc, "deladmin@test.com", "Admin")
	target := registerUser(t, userSvc, "deltarget@test.com", "Target")

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/"+target.ID.String(), nil)
	req = withClaims(req, admin.ID, domain.RoleAdmin)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", target.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.DeleteUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	h, userSvc, _ := newTestUserHandler()
	_ = registerUser(t, userSvc, "list1@test.com", "User One")
	_ = registerUser(t, userSvc, "list2@test.com", "User Two")

	req := httptest.NewRequest(http.MethodGet, "/admin/users?page=1&per_page=10", nil)
	admin := registerUser(t, userSvc, "listadmin@test.com", "Admin")
	req = withClaims(req, admin.ID, domain.RoleAdmin)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
