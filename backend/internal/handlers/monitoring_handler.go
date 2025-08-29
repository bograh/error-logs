package handlers

import (
	"net/http"

	"error-logs/internal/services"
)

type MonitoringHandler struct {
	monitoringService *services.MonitoringService
}

func NewMonitoringHandler(monitoringService *services.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

func (h *MonitoringHandler) GetServiceHealth(w http.ResponseWriter, r *http.Request) {
	services, err := h.monitoringService.GetServiceHealth(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get service health", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, services)
}

func (h *MonitoringHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	timeframe := r.URL.Query().Get("timeframe")
	if timeframe == "" {
		timeframe = "1h"
	}

	metrics, err := h.monitoringService.GetSystemMetrics(r.Context(), timeframe)
	if err != nil {
		writeErrorResponse(w, "Failed to get system metrics", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, metrics)
}

func (h *MonitoringHandler) GetUptime(w http.ResponseWriter, r *http.Request) {
	uptime, err := h.monitoringService.GetUptime(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to get uptime data", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, uptime)
}
