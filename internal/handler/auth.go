package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Rajneesh180/finance-backend/internal/api"
	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/service"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	userSvc  *service.UserService
	validate *validator.Validate
}

func NewAuthHandler(userSvc *service.UserService) *AuthHandler {
	return &AuthHandler{
		userSvc:  userSvc,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	user, err := h.userSvc.Register(r.Context(), req)
	if err != nil {
		if err == service.ErrEmailTaken {
			api.Conflict(w, "email already registered")
			return
		}
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
	if err := h.validate.Struct(req); err != nil {
		api.BadRequest(w, err.Error())
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
