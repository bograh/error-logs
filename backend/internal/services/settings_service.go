package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"error-logs/internal/database"
	"error-logs/internal/models"
	"error-logs/internal/redis"
)

type SettingsService struct {
	db    *database.DB
	redis *redis.Client
}

func NewSettingsService(db *database.DB, redis *redis.Client) *SettingsService {
	return &SettingsService{
		db:    db,
		redis: redis,
	}
}

func (s *SettingsService) GetAPIKeys(ctx context.Context) ([]models.APIKey, error) {
	return s.db.GetAPIKeys()
}

func (s *SettingsService) CreateAPIKey(ctx context.Context, req *models.CreateAPIKeyRequest, keyHash string) (*models.APIKey, error) {
	now := time.Now().UTC()

	apiKey := &models.APIKey{
		ID:          uuid.New(),
		KeyHash:     keyHash,
		Name:        req.Name,
		Permissions: req.Permissions,
		Active:      true,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   now,
		LastUsed:    nil,
	}

	if err := s.db.CreateAPIKey(apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *SettingsService) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	return s.db.DeleteAPIKey(id)
}

func (s *SettingsService) GetTeamMembers(ctx context.Context) ([]models.TeamMember, error) {
	return s.db.GetTeamMembers()
}

func (s *SettingsService) InviteTeamMember(ctx context.Context, req *models.InviteTeamMemberRequest) (*models.TeamMember, error) {
	now := time.Now().UTC()

	member := &models.TeamMember{
		ID:         uuid.New(),
		Name:       req.Email, // For now, use email as name
		Email:      req.Email,
		Role:       req.Role,
		Status:     "invited",
		LastActive: nil,
		CreatedAt:  now,
	}

	if err := s.db.CreateTeamMember(member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *SettingsService) GetIntegrations(ctx context.Context) ([]models.Integration, error) {
	// For demo purposes, return mock integrations
	// In a real implementation, this would be stored in database
	integrations := []models.Integration{
		{
			Name:   "slack",
			Status: "connected",
			Config: map[string]interface{}{
				"webhook_url": "https://hooks.slack.com/...",
				"channel":     "#alerts",
			},
			LastTest: &time.Time{}, // Mock last test time
		},
		{
			Name:   "email",
			Status: "configured",
			Config: map[string]interface{}{
				"smtp_server": "smtp.example.com",
				"from_email":  "alerts@example.com",
			},
			LastTest: nil,
		},
	}

	return integrations, nil
}
