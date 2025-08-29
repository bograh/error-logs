package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"error-logs/internal/models"
)

type DB struct {
	*sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

func (db *DB) CreateError(error *models.Error) error {
	query := `
		INSERT INTO errors (
			id, timestamp, level, message, stack_trace, context, source, 
			environment, user_agent, ip_address, url, fingerprint, resolved, 
			count, first_seen, last_seen, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)`

	contextJSON, err := json.Marshal(error.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	_, err = db.Exec(query,
		error.ID, error.Timestamp, error.Level, error.Message, error.StackTrace,
		contextJSON, error.Source, error.Environment, error.UserAgent,
		error.IPAddress, error.URL, error.Fingerprint, error.Resolved,
		error.Count, error.FirstSeen, error.LastSeen, error.CreatedAt, error.UpdatedAt,
	)

	return err
}

func (db *DB) GetErrors(limit, offset int, level, source string) ([]models.Error, int, error) {
	var errors []models.Error
	var total int

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if level != "" {
		whereClause += fmt.Sprintf(" AND level = $%d", argIndex)
		args = append(args, level)
		argIndex++
	}

	if source != "" {
		whereClause += fmt.Sprintf(" AND source = $%d", argIndex)
		args = append(args, source)
		argIndex++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM errors %s", whereClause)
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get errors
	query := fmt.Sprintf(`
		SELECT id, timestamp, level, message, stack_trace, context, source, 
			   environment, user_agent, ip_address, url, fingerprint, resolved, 
			   count, first_seen, last_seen, created_at, updated_at
		FROM errors %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query errors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Error
		var contextJSON []byte

		err := rows.Scan(
			&e.ID, &e.Timestamp, &e.Level, &e.Message, &e.StackTrace,
			&contextJSON, &e.Source, &e.Environment, &e.UserAgent,
			&e.IPAddress, &e.URL, &e.Fingerprint, &e.Resolved,
			&e.Count, &e.FirstSeen, &e.LastSeen, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan error: %w", err)
		}

		if err := json.Unmarshal(contextJSON, &e.Context); err != nil {
			e.Context = make(map[string]interface{})
		}

		errors = append(errors, e)
	}

	return errors, total, nil
}

func (db *DB) GetErrorByID(id uuid.UUID) (*models.Error, error) {
	query := `
		SELECT id, timestamp, level, message, stack_trace, context, source, 
			   environment, user_agent, ip_address, url, fingerprint, resolved, 
			   count, first_seen, last_seen, created_at, updated_at
		FROM errors WHERE id = $1
	`

	var e models.Error
	var contextJSON []byte

	err := db.QueryRow(query, id).Scan(
		&e.ID, &e.Timestamp, &e.Level, &e.Message, &e.StackTrace,
		&contextJSON, &e.Source, &e.Environment, &e.UserAgent,
		&e.IPAddress, &e.URL, &e.Fingerprint, &e.Resolved,
		&e.Count, &e.FirstSeen, &e.LastSeen, &e.CreatedAt, &e.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("error not found")
		}
		return nil, fmt.Errorf("failed to get error: %w", err)
	}

	if err := json.Unmarshal(contextJSON, &e.Context); err != nil {
		e.Context = make(map[string]interface{})
	}

	return &e, nil
}

func (db *DB) ResolveError(id uuid.UUID) error {
	query := "UPDATE errors SET resolved = true, updated_at = NOW() WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}

func (db *DB) DeleteError(id uuid.UUID) error {
	query := "DELETE FROM errors WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}

func (db *DB) GetStats() (*models.StatsResponse, error) {
	stats := &models.StatsResponse{}

	// Get total errors count
	err := db.QueryRow("SELECT COUNT(*) FROM errors").Scan(&stats.TotalErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to get total errors: %w", err)
	}

	// Get resolved errors count
	err = db.QueryRow("SELECT COUNT(*) FROM errors WHERE resolved = true").Scan(&stats.ResolvedErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to get resolved errors: %w", err)
	}

	// Get errors today count - using a more compatible date calculation
	err = db.QueryRow("SELECT COUNT(*) FROM errors WHERE DATE(timestamp) = CURRENT_DATE").Scan(&stats.ErrorsToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get errors today: %w", err)
	}

	// Get errors this week count - using a more compatible date calculation
	err = db.QueryRow("SELECT COUNT(*) FROM errors WHERE timestamp >= NOW() - INTERVAL '7 days'").Scan(&stats.ErrorsThisWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get errors this week: %w", err)
	}

	// Get errors this month count - using a more compatible date calculation
	err = db.QueryRow("SELECT COUNT(*) FROM errors WHERE timestamp >= NOW() - INTERVAL '30 days'").Scan(&stats.ErrorsThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get errors this month: %w", err)
	}

	// Calculate error rate for last 24 hours (errors per hour)
	var errors24h int
	err = db.QueryRow("SELECT COUNT(*) FROM errors WHERE timestamp >= NOW() - INTERVAL '24 hours'").Scan(&errors24h)
	if err == nil {
		stats.ErrorRate24h = float64(errors24h) / 24.0
	}

	// Calculate resolution rate
	if stats.TotalErrors > 0 {
		stats.ResolutionRate = (float64(stats.ResolvedErrors) / float64(stats.TotalErrors)) * 100
	}

	// Calculate average resolution time (mock for now)
	stats.AvgResolutionTime = "2h 15m"

	return stats, nil
}

func (db *DB) ValidateAPIKey(keyHash string) (*models.APIKey, error) {
	query := `
		SELECT id, key_hash, name, project_id, active, created_at, last_used
		FROM api_keys WHERE key_hash = $1 AND active = true
	`

	var apiKey models.APIKey
	err := db.QueryRow(query, keyHash).Scan(
		&apiKey.ID, &apiKey.KeyHash, &apiKey.Name, &apiKey.ProjectID,
		&apiKey.Active, &apiKey.CreatedAt, &apiKey.LastUsed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	// Update last used timestamp
	updateQuery := "UPDATE api_keys SET last_used = NOW() WHERE id = $1"
	db.Exec(updateQuery, apiKey.ID)

	return &apiKey, nil
}

// Analytics methods
func (db *DB) GetTrends(period, groupBy string) (*models.TrendResponse, error) {
	var timeFormat string

	// Determine time format based on groupBy
	switch groupBy {
	case "hour":
		timeFormat = "YYYY-MM-DD HH24:00:00"
	case "day":
		timeFormat = "YYYY-MM-DD"
	case "week":
		timeFormat = "YYYY-WW"
	case "month":
		timeFormat = "YYYY-MM"
	default:
		timeFormat = "YYYY-MM-DD"
	}

	// Determine the time range based on period
	var whereClause string
	switch period {
	case "day":
		whereClause = "WHERE timestamp >= NOW() - INTERVAL '24 hours'"
	case "week":
		whereClause = "WHERE timestamp >= NOW() - INTERVAL '7 days'"
	case "month":
		whereClause = "WHERE timestamp >= NOW() - INTERVAL '30 days'"
	case "year":
		whereClause = "WHERE timestamp >= NOW() - INTERVAL '1 year'"
	default:
		whereClause = "WHERE timestamp >= NOW() - INTERVAL '7 days'"
	}

	query := fmt.Sprintf(`
		SELECT 
			TO_CHAR(timestamp, '%s') as time_period,
			COUNT(*) as error_count,
			COUNT(CASE WHEN resolved = true THEN 1 END) as resolved_count,
			COUNT(CASE WHEN level = 'error' THEN 1 END) as critical_count
		FROM errors 
		%s
		GROUP BY time_period
		ORDER BY time_period ASC
	`, timeFormat, whereClause)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query trends: %w", err)
	}
	defer rows.Close()

	var dataPoints []models.TrendDataPoint
	for rows.Next() {
		var timePeriod string
		var errorCount, resolvedCount, criticalCount int

		err := rows.Scan(&timePeriod, &errorCount, &resolvedCount, &criticalCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trend data: %w", err)
		}

		// Parse the time period back to timestamp
		var timestamp time.Time
		switch groupBy {
		case "hour":
			timestamp, _ = time.Parse("2006-01-02 15:04:05", timePeriod)
		case "day":
			timestamp, _ = time.Parse("2006-01-02", timePeriod)
		case "week":
			timestamp, _ = time.Parse("2006-02", timePeriod) // Simplified for week
		case "month":
			timestamp, _ = time.Parse("2006-01", timePeriod)
		default:
			timestamp, _ = time.Parse("2006-01-02", timePeriod)
		}

		dataPoints = append(dataPoints, models.TrendDataPoint{
			Timestamp:     timestamp,
			ErrorCount:    errorCount,
			ResolvedCount: resolvedCount,
			CriticalCount: criticalCount,
		})
	}

	return &models.TrendResponse{
		Period:     period,
		DataPoints: dataPoints,
	}, nil
}

// Alert Rule methods
func (db *DB) GetAlertRules() ([]models.AlertRule, error) {
	query := `
		SELECT id, name, condition, threshold, time_window, enabled, 
			   notifications, last_triggered, created_at, updated_at
		FROM alert_rules ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		var notificationsJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Condition, &rule.Threshold,
			&rule.TimeWindow, &rule.Enabled, &notificationsJSON,
			&rule.LastTriggered, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		if err := json.Unmarshal(notificationsJSON, &rule.Notifications); err != nil {
			rule.Notifications = []string{}
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (db *DB) CreateAlertRule(rule *models.AlertRule) error {
	query := `
		INSERT INTO alert_rules (
			id, name, condition, threshold, time_window, enabled,
			notifications, last_triggered, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	notificationsJSON, err := json.Marshal(rule.Notifications)
	if err != nil {
		return fmt.Errorf("failed to marshal notifications: %w", err)
	}

	_, err = db.Exec(query,
		rule.ID, rule.Name, rule.Condition, rule.Threshold,
		rule.TimeWindow, rule.Enabled, notificationsJSON,
		rule.LastTriggered, rule.CreatedAt, rule.UpdatedAt,
	)

	return err
}

func (db *DB) GetAlertRuleByID(id uuid.UUID) (*models.AlertRule, error) {
	query := `
		SELECT id, name, condition, threshold, time_window, enabled,
			   notifications, last_triggered, created_at, updated_at
		FROM alert_rules WHERE id = $1
	`

	var rule models.AlertRule
	var notificationsJSON []byte

	err := db.QueryRow(query, id).Scan(
		&rule.ID, &rule.Name, &rule.Condition, &rule.Threshold,
		&rule.TimeWindow, &rule.Enabled, &notificationsJSON,
		&rule.LastTriggered, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert rule not found")
		}
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	if err := json.Unmarshal(notificationsJSON, &rule.Notifications); err != nil {
		rule.Notifications = []string{}
	}

	return &rule, nil
}

func (db *DB) UpdateAlertRule(rule *models.AlertRule) error {
	query := `
		UPDATE alert_rules SET 
			name = $2, condition = $3, threshold = $4, time_window = $5,
			enabled = $6, notifications = $7, updated_at = $8
		WHERE id = $1
	`

	notificationsJSON, err := json.Marshal(rule.Notifications)
	if err != nil {
		return fmt.Errorf("failed to marshal notifications: %w", err)
	}

	_, err = db.Exec(query,
		rule.ID, rule.Name, rule.Condition, rule.Threshold,
		rule.TimeWindow, rule.Enabled, notificationsJSON, rule.UpdatedAt,
	)

	return err
}

func (db *DB) DeleteAlertRule(id uuid.UUID) error {
	query := "DELETE FROM alert_rules WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}

// Incident methods
func (db *DB) GetIncidents() ([]models.Incident, error) {
	query := `
		SELECT id, title, severity, status, description, assigned_to, created_at, updated_at
		FROM incidents ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var incident models.Incident

		err := rows.Scan(
			&incident.ID, &incident.Title, &incident.Severity, &incident.Status,
			&incident.Description, &incident.AssignedTo, &incident.CreatedAt, &incident.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, incident)
	}

	return incidents, nil
}

func (db *DB) CreateIncident(incident *models.Incident) error {
	query := `
		INSERT INTO incidents (
			id, title, severity, status, description, assigned_to, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.Exec(query,
		incident.ID, incident.Title, incident.Severity, incident.Status,
		incident.Description, incident.AssignedTo, incident.CreatedAt, incident.UpdatedAt,
	)

	return err
}

func (db *DB) GetIncidentByID(id uuid.UUID) (*models.Incident, error) {
	query := `
		SELECT id, title, severity, status, description, assigned_to, created_at, updated_at
		FROM incidents WHERE id = $1
	`

	var incident models.Incident

	err := db.QueryRow(query, id).Scan(
		&incident.ID, &incident.Title, &incident.Severity, &incident.Status,
		&incident.Description, &incident.AssignedTo, &incident.CreatedAt, &incident.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("incident not found")
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return &incident, nil
}

func (db *DB) UpdateIncident(incident *models.Incident) error {
	query := `
		UPDATE incidents SET 
			title = $2, severity = $3, status = $4, description = $5,
			assigned_to = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := db.Exec(query,
		incident.ID, incident.Title, incident.Severity, incident.Status,
		incident.Description, incident.AssignedTo, incident.UpdatedAt,
	)

	return err
}

// API Key methods
func (db *DB) GetAPIKeys() ([]models.APIKey, error) {
	query := `
		SELECT id, key_hash, name, permissions, project_id, active, expires_at, created_at, last_used
		FROM api_keys WHERE active = true ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []models.APIKey
	for rows.Next() {
		var apiKey models.APIKey
		var permissionsJSON []byte

		err := rows.Scan(
			&apiKey.ID, &apiKey.KeyHash, &apiKey.Name, &permissionsJSON,
			&apiKey.ProjectID, &apiKey.Active, &apiKey.ExpiresAt,
			&apiKey.CreatedAt, &apiKey.LastUsed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
			apiKey.Permissions = []string{}
		}

		// Generate key preview (show first 4 and last 4 characters)
		if len(apiKey.KeyHash) >= 8 {
			apiKey.KeyPreview = "sk_****" + apiKey.KeyHash[len(apiKey.KeyHash)-4:]
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

func (db *DB) CreateAPIKey(apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (
			id, key_hash, name, permissions, project_id, active, expires_at, created_at, last_used
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	_, err = db.Exec(query,
		apiKey.ID, apiKey.KeyHash, apiKey.Name, permissionsJSON,
		apiKey.ProjectID, apiKey.Active, apiKey.ExpiresAt,
		apiKey.CreatedAt, apiKey.LastUsed,
	)

	return err
}

func (db *DB) DeleteAPIKey(id uuid.UUID) error {
	query := "UPDATE api_keys SET active = false WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}

// Team Member methods
func (db *DB) GetTeamMembers() ([]models.TeamMember, error) {
	query := `
		SELECT id, name, email, role, status, last_active, created_at
		FROM team_members ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember

		err := rows.Scan(
			&member.ID, &member.Name, &member.Email, &member.Role,
			&member.Status, &member.LastActive, &member.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		members = append(members, member)
	}

	return members, nil
}

func (db *DB) CreateTeamMember(member *models.TeamMember) error {
	query := `
		INSERT INTO team_members (
			id, name, email, role, status, last_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.Exec(query,
		member.ID, member.Name, member.Email, member.Role,
		member.Status, member.LastActive, member.CreatedAt,
	)

	return err
}
