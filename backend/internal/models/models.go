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
	TotalErrors       int     `json:"total_errors"`
	ResolvedErrors    int     `json:"resolved_errors"`
	ErrorsToday       int     `json:"errors_today"`
	ErrorsThisWeek    int     `json:"errors_this_week"`
	ErrorsThisMonth   int     `json:"errors_this_month"`
	ErrorRate24h      float64 `json:"error_rate_24h"`
	ResolutionRate    float64 `json:"resolution_rate"`
	AvgResolutionTime string  `json:"avg_resolution_time"`
}

// Analytics models
type TrendDataPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	ErrorCount    int       `json:"error_count"`
	ResolvedCount int       `json:"resolved_count"`
	CriticalCount int       `json:"critical_count"`
}

type TrendResponse struct {
	Period     string           `json:"period"`
	DataPoints []TrendDataPoint `json:"data_points"`
}

type PerformanceMetrics struct {
	AvgResponseTime     int     `json:"avg_response_time"`
	ErrorRatePercent    float64 `json:"error_rate_percent"`
	ThroughputRPM       int     `json:"throughput_rpm"`
	AvailabilityPercent float64 `json:"availability_percent"`
	PerformanceScore    float64 `json:"performance_score"`
}

// Monitoring models
type ServiceHealth struct {
	Name           string                 `json:"name"`
	Status         string                 `json:"status"`
	UptimePercent  float64                `json:"uptime_percent"`
	ResponseTimeMs int                    `json:"response_time_ms"`
	LastChecked    time.Time              `json:"last_checked"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

type ServicesResponse struct {
	Services      []ServiceHealth `json:"services"`
	OverallHealth string          `json:"overall_health"`
}

type SystemMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	DiskUsagePercent   float64 `json:"disk_usage_percent"`
	NetworkIO          struct {
		BytesIn  int64 `json:"bytes_in"`
		BytesOut int64 `json:"bytes_out"`
	} `json:"network_io"`
	ActiveConnections int `json:"active_connections"`
	RequestsPerMinute int `json:"requests_per_minute"`
}

type UptimeData struct {
	CurrentUptimeHours float64    `json:"current_uptime_hours"`
	UptimePercent24h   float64    `json:"uptime_percent_24h"`
	UptimePercent7d    float64    `json:"uptime_percent_7d"`
	UptimePercent30d   float64    `json:"uptime_percent_30d"`
	IncidentsCount     int        `json:"incidents_count"`
	LastDowntime       *time.Time `json:"last_downtime"`
}

// Alert models
type AlertRule struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	Condition     string     `json:"condition" db:"condition"`
	Threshold     int        `json:"threshold" db:"threshold"`
	TimeWindow    string     `json:"time_window" db:"time_window"`
	Enabled       bool       `json:"enabled" db:"enabled"`
	Notifications []string   `json:"notifications" db:"notifications"`
	LastTriggered *time.Time `json:"last_triggered" db:"last_triggered"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateAlertRuleRequest struct {
	Name          string   `json:"name"`
	Condition     string   `json:"condition"`
	Threshold     int      `json:"threshold"`
	TimeWindow    string   `json:"time_window"`
	Notifications []string `json:"notifications"`
	Enabled       bool     `json:"enabled"`
}

type Incident struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Severity    string     `json:"severity" db:"severity"`
	Status      string     `json:"status" db:"status"`
	Description string     `json:"description" db:"description"`
	AssignedTo  *uuid.UUID `json:"assigned_to" db:"assigned_to"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateIncidentRequest struct {
	Title       string     `json:"title"`
	Severity    string     `json:"severity"`
	Description string     `json:"description"`
	AssignedTo  *uuid.UUID `json:"assigned_to"`
}

// Settings models
type APIKey struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	KeyHash     string     `json:"-" db:"key_hash"`
	Name        string     `json:"name" db:"name"`
	KeyPreview  string     `json:"key_preview" db:"-"`
	Permissions []string   `json:"permissions" db:"permissions"`
	ProjectID   *uuid.UUID `json:"project_id" db:"project_id"`
	Active      bool       `json:"active" db:"active"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	LastUsed    *time.Time `json:"last_used" db:"last_used"`
}

type CreateAPIKeyRequest struct {
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type TeamMember struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	Email      string     `json:"email" db:"email"`
	Role       string     `json:"role" db:"role"`
	Status     string     `json:"status" db:"status"`
	LastActive *time.Time `json:"last_active" db:"last_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

type InviteTeamMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Integration struct {
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`
	Config   map[string]interface{} `json:"config"`
	LastTest *time.Time             `json:"last_test"`
}

// Response wrapper types
type APIResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
	Status string      `json:"status"`
}
