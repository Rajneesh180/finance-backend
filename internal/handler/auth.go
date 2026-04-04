package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Rajneesh180/finance-backend/internal/api"
	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/service"
)

type AuthHandler struct {
	userSvc *service.UserService
}

func NewAuthHandler(userSvc *service.UserService) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate.Struct(req); err != nil {
		api.BadRequest(w, api.ValidationErrors(err))
		return
	}

	user, err := h.userSvc.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			api.Conflict(w, "email already registered")
			return
		}
		slog.Error("register failed", "error", err)
		api.InternalError(w)
		return
	}

	api.JSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate.Struct(req); err != nil {
		api.BadRequest(w, api.ValidationErrors(err))
		return
	}

	token, user, err := h.userSvc.Login(r.Context(), req)
	if err != nil {
		api.Unauthorized(w, "invalid email or password")
		return
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}
