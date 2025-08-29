package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"error-logs/internal/models"
	"error-logs/internal/services"
)

type AlertsHandler struct {
	alertsService *services.AlertsService
}

func NewAlertsHandler(alertsService *services.AlertsService) *AlertsHandler {
	return &AlertsHandler{
		alertsService: alertsService,
	}
}

func (h *AlertsHandler) GetAlertRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.alertsService.GetAlertRules(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get alert rules", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, map[string]interface{}{"rules": rules})
}

func (h *AlertsHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		writeErrorResponse(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Condition == "" {
		writeErrorResponse(w, "Condition is required", http.StatusBadRequest)
		return
	}

	rule, err := h.alertsService.CreateAlertRule(r.Context(), &req)
	if err != nil {
		writeErrorResponse(w, "Failed to create alert rule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeSuccessResponse(w, rule)
}

func (h *AlertsHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErrorResponse(w, "Invalid alert rule ID", http.StatusBadRequest)
		return
	}

	var req models.CreateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	rule, err := h.alertsService.UpdateAlertRule(r.Context(), id, &req)
	if err != nil {
		writeErrorResponse(w, "Failed to update alert rule", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, rule)
}

func (h *AlertsHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErrorResponse(w, "Invalid alert rule ID", http.StatusBadRequest)
		return
	}

	err = h.alertsService.DeleteAlertRule(r.Context(), id)
	if err != nil {
		writeErrorResponse(w, "Failed to delete alert rule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AlertsHandler) GetIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.alertsService.GetIncidents(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get incidents", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, map[string]interface{}{"incidents": incidents})
}

func (h *AlertsHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" {
		writeErrorResponse(w, "Title is required", http.StatusBadRequest)
		return
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}

	incident, err := h.alertsService.CreateIncident(r.Context(), &req)
	if err != nil {
		writeErrorResponse(w, "Failed to create incident", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeSuccessResponse(w, incident)
}

func (h *AlertsHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErrorResponse(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	var req models.CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	incident, err := h.alertsService.UpdateIncident(r.Context(), id, &req)
	if err != nil {
		writeErrorResponse(w, "Failed to update incident", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, incident)
}
