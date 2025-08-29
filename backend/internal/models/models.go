package models

import (
	"time"

	"github.com/google/uuid"
)

type Error struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Level       string                 `json:"level" db:"level"`
	Message     string                 `json:"message" db:"message"`
	StackTrace  *string                `json:"stack_trace" db:"stack_trace"`
	Context     map[string]interface{} `json:"context" db:"context"`
	Source      string                 `json:"source" db:"source"`
	Environment string                 `json:"environment" db:"environment"`
	UserAgent   *string                `json:"user_agent" db:"user_agent"`
	IPAddress   *string                `json:"ip_address" db:"ip_address"`
	URL         *string                `json:"url" db:"url"`
	Fingerprint *string                `json:"fingerprint" db:"fingerprint"`
	Resolved    bool                   `json:"resolved" db:"resolved"`
	Count       int                    `json:"count" db:"count"`
	FirstSeen   time.Time              `json:"first_seen" db:"first_seen"`
	LastSeen    time.Time              `json:"last_seen" db:"last_seen"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type CreateErrorRequest struct {
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	StackTrace  *string                `json:"stack_trace"`
	Context     map[string]interface{} `json:"context"`
	Source      string                 `json:"source"`
	Environment *string                `json:"environment"`
	URL         *string                `json:"url"`
}

type ErrorListResponse struct {
	Errors []Error `json:"errors"`
	Total  int     `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

type StatsResponse struct {
	TotalErrors     int `json:"total_errors"`
	ResolvedErrors  int `json:"resolved_errors"`
	ErrorsToday     int `json:"errors_today"`
	ErrorsThisWeek  int `json:"errors_this_week"`
	ErrorsThisMonth int `json:"errors_this_month"`
}

type APIKey struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	KeyHash   string     `json:"-" db:"key_hash"`
	Name      string     `json:"name" db:"name"`
	ProjectID *uuid.UUID `json:"project_id" db:"project_id"`
	Active    bool       `json:"active" db:"active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	LastUsed  *time.Time `json:"last_used" db:"last_used"`
}
