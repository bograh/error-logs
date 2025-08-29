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
