package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	log     *slog.Logger
}

func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	r.Post("/refresh", h.refresh)
	return r
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDuplicate) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		h.log.ErrorContext(r.Context(), "register failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.log.ErrorContext(r.Context(), "login failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.service.Logout(r.Context(), body.RefreshToken); err != nil {
		h.log.ErrorContext(r.Context(), "logout failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.service.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		h.log.ErrorContext(r.Context(), "refresh failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
