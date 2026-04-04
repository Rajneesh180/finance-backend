package handler

import (
	"net/http"
	"strconv"

	"github.com/Rajneesh180/finance-backend/internal/api"
	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/middleware"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	recordSvc *service.RecordService
}

func NewDashboardHandler(recordSvc *service.RecordService) *DashboardHandler {
	return &DashboardHandler{recordSvc: recordSvc}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	q := r.URL.Query()

	var userID *uuid.UUID
	if claims.Role != domain.RoleAdmin {
		userID = &claims.UserID
	}

	summary, err := h.recordSvc.Dashboard(r.Context(), userID, q.Get("date_from"), q.Get("date_to"))
	if err != nil {
		api.InternalError(w)
		return
	}

	api.JSON(w, http.StatusOK, summary)
}

func (h *DashboardHandler) RecentActivity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var userID *uuid.UUID
	if claims.Role != domain.RoleAdmin {
		userID = &claims.UserID
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	entries, err := h.recordSvc.RecentActivity(r.Context(), userID, limit)
	if err != nil {
		api.InternalError(w)
		return
	}

	api.JSON(w, http.StatusOK, entries)
}
