package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"error-logs/internal/database"
	"error-logs/internal/models"
	"error-logs/internal/redis"
)

type AlertsService struct {
	db    *database.DB
	redis *redis.Client
}

func NewAlertsService(db *database.DB, redis *redis.Client) *AlertsService {
	return &AlertsService{
		db:    db,
		redis: redis,
	}
}

func (s *AlertsService) GetAlertRules(ctx context.Context) ([]models.AlertRule, error) {
	return s.db.GetAlertRules()
}

func (s *AlertsService) CreateAlertRule(ctx context.Context, req *models.CreateAlertRuleRequest) (*models.AlertRule, error) {
	now := time.Now().UTC()

	rule := &models.AlertRule{
		ID:            uuid.New(),
		Name:          req.Name,
		Condition:     req.Condition,
		Threshold:     req.Threshold,
		TimeWindow:    req.TimeWindow,
		Enabled:       req.Enabled,
		Notifications: req.Notifications,
		LastTriggered: nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.db.CreateAlertRule(rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *AlertsService) UpdateAlertRule(ctx context.Context, id uuid.UUID, req *models.CreateAlertRuleRequest) (*models.AlertRule, error) {
	rule, err := s.db.GetAlertRuleByID(id)
	if err != nil {
		return nil, err
	}

	rule.Name = req.Name
	rule.Condition = req.Condition
	rule.Threshold = req.Threshold
	rule.TimeWindow = req.TimeWindow
	rule.Enabled = req.Enabled
	rule.Notifications = req.Notifications
	rule.UpdatedAt = time.Now().UTC()

	if err := s.db.UpdateAlertRule(rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *AlertsService) DeleteAlertRule(ctx context.Context, id uuid.UUID) error {
	return s.db.DeleteAlertRule(id)
}

func (s *AlertsService) GetIncidents(ctx context.Context) ([]models.Incident, error) {
	return s.db.GetIncidents()
}

func (s *AlertsService) CreateIncident(ctx context.Context, req *models.CreateIncidentRequest) (*models.Incident, error) {
	now := time.Now().UTC()

	incident := &models.Incident{
		ID:          uuid.New(),
		Title:       req.Title,
		Severity:    req.Severity,
		Status:      "open",
		Description: req.Description,
		AssignedTo:  req.AssignedTo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.db.CreateIncident(incident); err != nil {
		return nil, err
	}

	return incident, nil
}

func (s *AlertsService) UpdateIncident(ctx context.Context, id uuid.UUID, req *models.CreateIncidentRequest) (*models.Incident, error) {
	incident, err := s.db.GetIncidentByID(id)
	if err != nil {
		return nil, err
	}

	incident.Title = req.Title
	incident.Severity = req.Severity
	incident.Description = req.Description
	incident.AssignedTo = req.AssignedTo
	incident.UpdatedAt = time.Now().UTC()

	if err := s.db.UpdateIncident(incident); err != nil {
		return nil, err
	}

	return incident, nil
}
