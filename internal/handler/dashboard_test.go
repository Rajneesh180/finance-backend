package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/Rajneesh180/finance-backend/internal/testutil"
	"github.com/google/uuid"
)

func newTestDashboardHandler() (*DashboardHandler, uuid.UUID) {
	repo := testutil.NewMockRecordRepo()
	dash := testutil.NewMockDashboardRepo()
	recordSvc := service.NewRecordService(repo, dash)
	userID := uuid.New()
	return NewDashboardHandler(recordSvc), userID
}

func TestDashboardHandler_Summary(t *testing.T) {
	h, userID := newTestDashboardHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
	req = withClaims(req, userID, domain.RoleAnalyst)
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["total_income"] != "5000" {
		t.Errorf("got total_income %v, want 5000", data["total_income"])
	}
}

func TestDashboardHandler_Summary_WithDateRange(t *testing.T) {
	h, userID := newTestDashboardHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/summary?date_from=2026-01-01&date_to=2026-12-31", nil)
	req = withClaims(req, userID, domain.RoleViewer)
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestDashboardHandler_Summary_AdminSeesAll(t *testing.T) {
	h, userID := newTestDashboardHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
	req = withClaims(req, userID, domain.RoleAdmin)
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestDashboardHandler_RecentActivity(t *testing.T) {
	h, userID := newTestDashboardHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/recent?limit=5", nil)
	req = withClaims(req, userID, domain.RoleAnalyst)
	rr := httptest.NewRecorder()

	h.RecentActivity(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["data"] == nil {
		t.Error("expected data field in response")
	}
}

func TestDashboardHandler_RecentActivity_DefaultLimit(t *testing.T) {
	h, userID := newTestDashboardHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/recent", nil)
	req = withClaims(req, userID, domain.RoleViewer)
	rr := httptest.NewRecorder()

	h.RecentActivity(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}
}
