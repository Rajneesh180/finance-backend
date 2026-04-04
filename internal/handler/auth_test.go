package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/Rajneesh180/finance-backend/internal/testutil"
)

func newTestAuthHandler() *AuthHandler {
	repo := testutil.NewMockUserRepo()
	auth := service.NewAuthService("handler-test-secret", 60)
	userSvc := service.NewUserService(repo, auth)
	return NewAuthHandler(userSvc)
}

func TestAuthHandler_Register(t *testing.T) {
	h := newTestAuthHandler()

	body := `{"email":"new@example.com","password":"password123","name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field in response")
	}
	if data["email"] != "new@example.com" {
		t.Errorf("got email %v, want new@example.com", data["email"])
	}
}

func TestAuthHandler_Register_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{bad json`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	h := newTestAuthHandler()

	body := `{"email":"a@b.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	h := newTestAuthHandler()

	body := `{"email":"dup@example.com","password":"password123","name":"First"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first register failed: %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.Register(rr2, req2)

	if rr2.Code != http.StatusConflict {
		t.Errorf("got status %d, want %d", rr2.Code, http.StatusConflict)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	h := newTestAuthHandler()

	regBody := `{"email":"login@example.com","password":"password123","name":"Login User"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	h.Register(regRR, regReq)

	loginBody := `{"email":"login@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["token"] == nil || data["token"] == "" {
		t.Error("expected token in login response")
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	h := newTestAuthHandler()

	regBody := `{"email":"wrong@example.com","password":"password123","name":"User"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	h.Register(httptest.NewRecorder(), regReq)

	loginBody := `{"email":"wrong@example.com","password":"badpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}
