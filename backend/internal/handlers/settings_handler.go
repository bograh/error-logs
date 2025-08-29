package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"error-logs/internal/models"
	"error-logs/internal/services"
)

type SettingsHandler struct {
	settingsService *services.SettingsService
}

func NewSettingsHandler(settingsService *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

func (h *SettingsHandler) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	apiKeys, err := h.settingsService.GetAPIKeys(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get API keys", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, map[string]interface{}{"api_keys": apiKeys})
}

func (h *SettingsHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		writeErrorResponse(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Permissions) == 0 {
		req.Permissions = []string{"read"}
	}

	// Generate API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		writeErrorResponse(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	apiKey := "sk_" + hex.EncodeToString(keyBytes)
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(apiKey)))

	key, err := h.settingsService.CreateAPIKey(r.Context(), &req, keyHash)
	if err != nil {
		writeErrorResponse(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Return the key with the actual API key (only time it's shown)
	response := map[string]interface{}{
		"id":          key.ID,
		"name":        key.Name,
		"api_key":     apiKey, // Only shown once
		"permissions": key.Permissions,
		"expires_at":  key.ExpiresAt,
		"created_at":  key.CreatedAt,
	}

	w.WriteHeader(http.StatusCreated)
	writeSuccessResponse(w, response)
}

func (h *SettingsHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErrorResponse(w, "Invalid API key ID", http.StatusBadRequest)
		return
	}

	err = h.settingsService.DeleteAPIKey(r.Context(), id)
	if err != nil {
		writeErrorResponse(w, "Failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SettingsHandler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.settingsService.GetTeamMembers(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get team members", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, map[string]interface{}{"members": members})
}

func (h *SettingsHandler) InviteTeamMember(w http.ResponseWriter, r *http.Request) {
	var req models.InviteTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" {
		writeErrorResponse(w, "Email is required", http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "viewer"
	}

	member, err := h.settingsService.InviteTeamMember(r.Context(), &req)
	if err != nil {
		writeErrorResponse(w, "Failed to invite team member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeSuccessResponse(w, member)
}

func (h *SettingsHandler) GetIntegrations(w http.ResponseWriter, r *http.Request) {
	integrations, err := h.settingsService.GetIntegrations(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get integrations", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, map[string]interface{}{"integrations": integrations})
}
