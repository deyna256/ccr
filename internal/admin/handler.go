package admin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Storage interface {
	SetAdmin(ctx context.Context, userID string, isAdmin bool) error
	ListUserStats(ctx context.Context) ([]UserStats, error)
}

type Handler struct {
	storage      Storage
	secret       string
	log          *slog.Logger
	cookieSecret []byte
}

func NewHandler(storage Storage, secret string, log *slog.Logger) *Handler {
	return &Handler{
		storage:      storage,
		secret:       secret,
		log:          log,
		cookieSecret: []byte(secret + "-admin"),
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/api/login", h.handleLogin)
	r.Post("/api/logout", h.handleLogout)
	r.Get("/api/users", h.handleUsers)
	r.Post("/api/users/{userID}/makeadmin", h.handleMakeAdmin)
	r.Post("/api/users/{userID}/unadmin", h.handleUnadmin)
	r.Get("/", h.serveLoginPage)
	r.Get("/login", h.serveLoginPage)
	r.Get("/dashboard", h.serveDashboard)
	return r
}

type LoginRequest struct {
	Secret string `json:"secret"`
}

const cookieMaxAgeSeconds = 86400

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Secret != h.secret {
		h.log.WarnContext(r.Context(), "admin login failed: invalid secret")
		writeError(w, http.StatusUnauthorized, "invalid secret")
		return
	}

	token := generateToken(h.cookieSecret)
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   cookieMaxAgeSeconds,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true}) //nolint:errcheck
}

func (h *Handler) handleLogout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true}) //nolint:errcheck
}

func (h *Handler) handleUsers(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r, h.cookieSecret) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()
	users, err := h.storage.ListUserStats(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "list users failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]UserStats{"users": users}) //nolint:errcheck
}

func (h *Handler) handleMakeAdmin(w http.ResponseWriter, r *http.Request) {
	h.setAdmin(w, r, true)
}

func (h *Handler) handleUnadmin(w http.ResponseWriter, r *http.Request) {
	h.setAdmin(w, r, false)
}

func (h *Handler) setAdmin(w http.ResponseWriter, r *http.Request, isAdmin bool) {
	if !checkAuth(r, h.cookieSecret) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	ctx := r.Context()
	err := h.storage.SetAdmin(ctx, userID, isAdmin)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.ErrorContext(ctx, "set admin failed", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	action := "makeadmin"
	if !isAdmin {
		action = "unadmin"
	}
	h.log.InfoContext(ctx, action, slog.String("user_id", userID))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true}) //nolint:errcheck
}

func checkAuth(r *http.Request, secret []byte) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return false
	}
	return validateToken(cookie.Value, secret)
}

func generateToken(secret []byte) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(timestamp))
	return timestamp + "." + base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func validateToken(token string, secret []byte) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	expected := hmac.New(sha256.New, secret)
	expected.Write([]byte(parts[0]))
	decoded, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	return hmac.Equal(decoded, expected.Sum(nil))
}

func (h *Handler) serveLoginPage(w http.ResponseWriter, r *http.Request) {
	if checkAuth(r, h.cookieSecret) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	serveHTML(w, r, "login.html")
}

func (h *Handler) serveDashboard(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r, h.cookieSecret) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	serveHTML(w, r, "dashboard.html")
}

func serveHTML(w http.ResponseWriter, _ *http.Request, filename string) {
	uiFS, err := fs.Sub(adminUI, "ui")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	f, err := uiFS.Open(filename)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, f) //nolint:errcheck
	f.Close()     //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
