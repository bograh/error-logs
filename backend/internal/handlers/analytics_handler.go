package handlers

import (
	"encoding/json"
	"net/http"

	"error-logs/internal/models"
	"error-logs/internal/services"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}

	trends, err := h.analyticsService.GetTrends(r.Context(), period, groupBy)
	if err != nil {
		writeErrorResponse(w, "Failed to get trends", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, trends)
}

func (h *AnalyticsHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.analyticsService.GetPerformanceMetrics(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get performance metrics", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, metrics)
}

// Helper functions
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := models.APIResponse{
		Data:   data,
		Status: "success",
	}
	json.NewEncoder(w).Encode(response)
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.APIResponse{
		Error:  message,
		Status: "error",
	}
	json.NewEncoder(w).Encode(response)
}
