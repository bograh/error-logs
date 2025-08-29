package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"error-logs/internal/database"
	"error-logs/internal/models"
	"error-logs/internal/services"
)

type ErrorHandler struct {
	errorService *services.ErrorService
}

func NewErrorHandler(errorService *services.ErrorService) *ErrorHandler {
	return &ErrorHandler{
		errorService: errorService,
	}
}

// APIKeyMiddleware validates API keys
func APIKeyMiddleware(db *database.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Hash the API key
			hash := sha256.Sum256([]byte(apiKey))
			keyHash := fmt.Sprintf("%x", hash)

			_, err := db.ValidateAPIKey(keyHash)
			if err != nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (h *ErrorHandler) CreateError(w http.ResponseWriter, r *http.Request) {
	var req models.CreateErrorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	if req.Level == "" {
		req.Level = "error"
	}
	if req.Source == "" {
		req.Source = "unknown"
	}

	// Extract client info
	userAgent := r.Header.Get("User-Agent")
	ipAddress := getClientIP(r)

	error, err := h.errorService.CreateError(r.Context(), &req, userAgent, ipAddress)
	if err != nil {
		http.Error(w, "Failed to create error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(error)
}

func (h *ErrorHandler) GetErrors(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	level := r.URL.Query().Get("level")
	source := r.URL.Query().Get("source")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	response, err := h.errorService.GetErrors(r.Context(), limit, offset, level, source)
	if err != nil {
		http.Error(w, "Failed to get errors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ErrorHandler) GetError(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid error ID", http.StatusBadRequest)
		return
	}

	error, err := h.errorService.GetErrorByID(r.Context(), id)
	if err != nil {
		if err.Error() == "error not found" {
			http.Error(w, "Error not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(error)
}

func (h *ErrorHandler) ResolveError(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid error ID", http.StatusBadRequest)
		return
	}

	err = h.errorService.ResolveError(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to resolve error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

func (h *ErrorHandler) DeleteError(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid error ID", http.StatusBadRequest)
		return
	}

	err = h.errorService.DeleteError(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ErrorHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.errorService.GetStats(r.Context())
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	if stats == nil {
		log.Printf("Stats returned nil")
		http.Error(w, "No stats available", http.StatusInternalServerError)
		return
	}

	log.Printf("Returning stats to client: %+v", stats)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Failed to encode stats: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr, but strip the port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If we can't split host:port, return as-is (might be just an IP)
		return r.RemoteAddr
	}
	return host
}
