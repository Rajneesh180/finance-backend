package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/middleware"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/Rajneesh180/finance-backend/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newTestRecordHandler() (*RecordHandler, uuid.UUID) {
	repo := testutil.NewMockRecordRepo()
	dash := testutil.NewMockDashboardRepo()
	recordSvc := service.NewRecordService(repo, dash)
	userID := uuid.New()
	return NewRecordHandler(recordSvc), userID
}

func withClaims(r *http.Request, userID uuid.UUID, role domain.Role) *http.Request {
	claims := &service.Claims{
		UserID: userID,
		Email:  "test@example.com",
		Role:   role,
	}
	ctx := context.WithValue(r.Context(), middleware.ClaimsKey, claims)
	return r.WithContext(ctx)
}

func TestRecordHandler_Create(t *testing.T) {
	h, userID := newTestRecordHandler()

	body := `{"amount":"500.00","type":"income","category":"salary","date":"2026-03-15"}`
	req := httptest.NewRequest(http.MethodPost, "/records", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, userID, domain.RoleAnalyst)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

func TestRecordHandler_Create_InvalidAmount(t *testing.T) {
	h, userID := newTestRecordHandler()

	body := `{"amount":"abc","type":"income","category":"salary","date":"2026-03-15"}`
	req := httptest.NewRequest(http.MethodPost, "/records", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, userID, domain.RoleAnalyst)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRecordHandler_GetByID(t *testing.T) {
	h, userID := newTestRecordHandler()

	// Create first
	body := `{"amount":"250.00","type":"expense","category":"food","date":"2026-03-10"}`
	createReq := httptest.NewRequest(http.MethodPost, "/records", bytes.NewBufferString(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withClaims(createReq, userID, domain.RoleAnalyst)
	createRR := httptest.NewRecorder()
	h.Create(createRR, createReq)

	var createResp map[string]interface{}
	json.NewDecoder(createRR.Body).Decode(&createResp)
	data := createResp["data"].(map[string]interface{})
	recordID := data["id"].(string)

	// Now get by ID via chi route context
	getReq := httptest.NewRequest(http.MethodGet, "/records/"+recordID, nil)
	getReq = withClaims(getReq, userID, domain.RoleAnalyst)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", recordID)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rctx))

	getRR := httptest.NewRecorder()
	h.GetByID(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", getRR.Code, http.StatusOK, getRR.Body.String())
	}
}

func TestRecordHandler_List(t *testing.T) {
	h, userID := newTestRecordHandler()

	// Create a couple records
	for _, amt := range []string{"100", "200"} {
		body := `{"amount":"` + amt + `","type":"expense","category":"food","date":"2026-03-10"}`
		req := httptest.NewRequest(http.MethodPost, "/records", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withClaims(req, userID, domain.RoleAnalyst)
		h.Create(httptest.NewRecorder(), req)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/records?page=1&per_page=10", nil)
	listReq = withClaims(listReq, userID, domain.RoleAnalyst)
	listRR := httptest.NewRecorder()
	h.List(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", listRR.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(listRR.Body).Decode(&resp)
	meta := resp["meta"].(map[string]interface{})
	total := int(meta["total"].(float64))
	if total != 2 {
		t.Errorf("got total %d, want 2", total)
	}
}

func TestRecordHandler_Delete_NotFound(t *testing.T) {
	h, userID := newTestRecordHandler()

	fakeID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/records/"+fakeID, nil)
	req = withClaims(req, userID, domain.RoleAnalyst)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fakeID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusNotFound)
	}
}
