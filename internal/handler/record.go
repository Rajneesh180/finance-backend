package handler

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/Rajneesh180/finance-backend/internal/api"
	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/middleware"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RecordHandler struct {
	recordSvc *service.RecordService
}

func NewRecordHandler(recordSvc *service.RecordService) *RecordHandler {
	return &RecordHandler{recordSvc: recordSvc}
}

func (h *RecordHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req domain.CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate.Struct(req); err != nil {
		api.BadRequest(w, api.ValidationErrors(err))
		return
	}

	record, err := h.recordSvc.Create(r.Context(), claims.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			api.BadRequest(w, "amount must be a positive number")
		case errors.Is(err, service.ErrInvalidDate):
			api.BadRequest(w, err.Error())
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusCreated, record)
}

func (h *RecordHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid record id")
		return
	}

	isAdmin := claims.Role == domain.RoleAdmin
	record, err := h.recordSvc.GetByID(r.Context(), id, claims.UserID, isAdmin)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			api.NotFound(w, "record not found")
		case errors.Is(err, service.ErrNotOwner):
			api.Forbidden(w, "not your record")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusOK, record)
}

func (h *RecordHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid record id")
		return
	}

	var req domain.UpdateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate.Struct(req); err != nil {
		api.BadRequest(w, api.ValidationErrors(err))
		return
	}

	isAdmin := claims.Role == domain.RoleAdmin
	record, err := h.recordSvc.Update(r.Context(), id, claims.UserID, isAdmin, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			api.NotFound(w, "record not found")
		case errors.Is(err, service.ErrNotOwner):
			api.Forbidden(w, "not your record")
		case errors.Is(err, service.ErrInvalidAmount):
			api.BadRequest(w, "amount must be a positive number")
		case errors.Is(err, service.ErrInvalidDate):
			api.BadRequest(w, err.Error())
		case errors.Is(err, service.ErrInvalidRecordType):
			api.BadRequest(w, "invalid record type")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusOK, record)
}

func (h *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid record id")
		return
	}

	isAdmin := claims.Role == domain.RoleAdmin
	if err := h.recordSvc.Delete(r.Context(), id, claims.UserID, isAdmin); err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			api.NotFound(w, "record not found")
		case errors.Is(err, service.ErrNotOwner):
			api.Forbidden(w, "not your record")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{"message": "record deleted"})
}

func (h *RecordHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	q := r.URL.Query()

	filter := domain.RecordFilter{
		Type:      q.Get("type"),
		Category:  q.Get("category"),
		DateFrom:  q.Get("date_from"),
		DateTo:    q.Get("date_to"),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	filter.Page = page
	filter.PerPage = perPage

	var userID *uuid.UUID
	if claims.Role != domain.RoleAdmin {
		userID = &claims.UserID
	}

	records, total, err := h.recordSvc.List(r.Context(), userID, filter)
	if err != nil {
		api.InternalError(w)
		return
	}

	meta := domain.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	}
	api.JSONWithMeta(w, http.StatusOK, records, meta)
}
