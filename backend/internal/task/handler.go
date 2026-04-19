package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
	r.Get("/{id}", h.getByID)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Patch("/{id}/status", h.updateStatus)
	r.Get("/{id}/attachments", h.listAttachments)
	r.Post("/{id}/attachments", h.uploadAttachment)
	r.Delete("/{id}/attachments/{aid}", h.deleteAttachment)
	return r
}

func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	a, err := h.service.GetAttachment(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "attachment not found")
			return
		}
		h.log.ErrorContext(r.Context(), "get attachment failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", a.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", a.Name))
	http.ServeFile(w, r, a.FilePath)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var f ListFilter
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid from parameter")
			return
		}
		f.From = &t
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid to parameter")
			return
		}
		f.To = &t
	}
	tasks, err := h.service.List(r.Context(), userID, f)
	if err != nil {
		h.log.ErrorContext(r.Context(), "list tasks failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if tasks == nil {
		tasks = []Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	t, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		h.log.ErrorContext(r.Context(), "create task failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	t, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		h.log.ErrorContext(r.Context(), "get task failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	t, err := h.service.Update(r.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, ErrInvalid) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		h.log.ErrorContext(r.Context(), "update task failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		h.log.ErrorContext(r.Context(), "delete task failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	t, err := h.service.UpdateStatus(r.Context(), id, userID, req.Status)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, ErrInvalid) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		h.log.ErrorContext(r.Context(), "update status failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) listAttachments(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	atts, err := h.service.ListAttachments(r.Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		h.log.ErrorContext(r.Context(), "list attachments failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if atts == nil {
		atts = []Attachment{}
	}
	writeJSON(w, http.StatusOK, atts)
}

func (h *Handler) uploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	a, err := h.service.UploadAttachment(r.Context(), taskID, userID, header.Filename, mimeType, header.Size, file)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		h.log.ErrorContext(r.Context(), "upload attachment failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	aid := chi.URLParam(r, "aid")
	if err := h.service.DeleteAttachment(r.Context(), aid, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "attachment not found")
			return
		}
		h.log.ErrorContext(r.Context(), "delete attachment failed", slog.String("error", err.Error()))
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
