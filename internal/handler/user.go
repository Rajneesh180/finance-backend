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

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	user, err := h.userSvc.GetByID(r.Context(), claims.UserID)
	if err != nil {
		api.NotFound(w, "user not found")
		return
	}
	api.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate.Struct(req); err != nil {
		api.BadRequest(w, api.ValidationErrors(err))
		return
	}

	user, err := h.userSvc.UpdateProfile(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			api.Conflict(w, "email already in use")
			return
		}
		api.InternalError(w)
		return
	}
	api.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid user id")
		return
	}
	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		api.NotFound(w, "user not found")
		return
	}
	api.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid user id")
		return
	}

	var req domain.AdminUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate.Struct(req); err != nil {
		api.BadRequest(w, api.ValidationErrors(err))
		return
	}

	user, err := h.userSvc.AdminUpdate(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			api.NotFound(w, "user not found")
		case errors.Is(err, service.ErrInvalidRole):
			api.BadRequest(w, "invalid role")
		default:
			api.InternalError(w)
		}
		return
	}
	api.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid user id")
		return
	}
	if err := h.userSvc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			api.NotFound(w, "user not found")
			return
		}
		api.InternalError(w)
		return
	}
	api.JSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	users, total, err := h.userSvc.List(r.Context(), page, perPage)
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
	api.JSONWithMeta(w, http.StatusOK, users, meta)
}
