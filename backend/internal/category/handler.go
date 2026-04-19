package category

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/task-planner/server/internal/auth"
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
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	cats, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "list categories failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if cats == nil {
		cats = []Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	cat, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		h.log.ErrorContext(r.Context(), "create category failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	cat, err := h.service.Update(r.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		if errors.Is(err, ErrInvalid) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		h.log.ErrorContext(r.Context(), "update category failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		h.log.ErrorContext(r.Context(), "delete category failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
