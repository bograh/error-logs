package services

import (
	"context"
	"log"
	"math/rand"
	"time"

	"error-logs/internal/database"
	"error-logs/internal/models"
	"error-logs/internal/redis"
)

type MonitoringService struct {
	db    *database.DB
	redis *redis.Client
}

func NewMonitoringService(db *database.DB, redis *redis.Client) *MonitoringService {
	return &MonitoringService{
		db:    db,
		redis: redis,
	}
}

func (s *MonitoringService) GetServiceHealth(ctx context.Context) (*models.ServicesResponse, error) {
	// Try to get from cache first
	if cachedHealth, err := s.redis.GetCachedServiceHealth(ctx); err == nil && cachedHealth != nil {
		log.Printf("CACHE HIT: GetServiceHealth")
		return cachedHealth, nil
	}

	log.Printf("CACHE MISS: GetServiceHealth - generating health data")

	// Check database health
	dbHealth := s.checkDatabaseHealth()

	// Check Redis health
	redisHealth := s.checkRedisHealth()

	// Mock additional services for demo
	apiHealth := models.ServiceHealth{
		Name:           "API Service",
		Status:         "healthy",
		UptimePercent:  99.95,
		ResponseTimeMs: 50 + rand.Intn(50),
		LastChecked:    time.Now().UTC(),
		Details: map[string]interface{}{
			"active_requests": rand.Intn(100),
			"max_requests":    1000,
		},
	}

	services := []models.ServiceHealth{dbHealth, redisHealth, apiHealth}

	// Determine overall health
	overallHealth := "healthy"
	for _, service := range services {
		if service.Status != "healthy" {
			overallHealth = "unhealthy"
			break
		}
	}

	response := &models.ServicesResponse{
		Services:      services,
		OverallHealth: overallHealth,
	}

	// Cache the result
	go func() {
		cacheCtx := context.Background()
		if err := s.redis.CacheServiceHealth(cacheCtx, response, 30*time.Second); err != nil {
			log.Printf("Failed to cache service health: %v", err)
		} else {
			log.Printf("CACHE WRITE: GetServiceHealth")
		}
	}()

	return response, nil
}

func (s *MonitoringService) checkDatabaseHealth() models.ServiceHealth {
	start := time.Now()

	err := s.db.Ping()
	responseTime := int(time.Since(start).Milliseconds())

	status := "healthy"
	if err != nil {
		status = "unhealthy"
		responseTime = 0
	}

	return models.ServiceHealth{
		Name:           "Database",
		Status:         status,
		UptimePercent:  99.8,
		ResponseTimeMs: responseTime,
		LastChecked:    time.Now().UTC(),
		Details: map[string]interface{}{
			"connections":     rand.Intn(50) + 10,
			"max_connections": 100,
		},
	}
}

func (s *MonitoringService) checkRedisHealth() models.ServiceHealth {
	start := time.Now()

	err := s.redis.Ping(context.Background()).Err()
	responseTime := int(time.Since(start).Milliseconds())

	status := "healthy"
	if err != nil {
		status = "unhealthy"
		responseTime = 0
	}

	return models.ServiceHealth{
		Name:           "Cache Service",
		Status:         status,
		UptimePercent:  99.95,
		ResponseTimeMs: responseTime,
		LastChecked:    time.Now().UTC(),
	}
}

func (s *MonitoringService) GetSystemMetrics(ctx context.Context, timeframe string) (*models.SystemMetrics, error) {
	// Try to get from cache first
	if cachedMetrics, err := s.redis.GetCachedSystemMetrics(ctx); err == nil && cachedMetrics != nil {
		log.Printf("CACHE HIT: GetSystemMetrics")
		return cachedMetrics, nil
	}

	log.Printf("CACHE MISS: GetSystemMetrics - generating mock metrics")

	// For demo purposes, generate mock system metrics
	// In a real implementation, this would collect actual system metrics
	metrics := &models.SystemMetrics{
		CPUUsagePercent:    65.2 + rand.Float64()*20,
		MemoryUsagePercent: 78.1 + rand.Float64()*10,
		DiskUsagePercent:   45.7 + rand.Float64()*10,
		NetworkIO: struct {
			BytesIn  int64 `json:"bytes_in"`
			BytesOut int64 `json:"bytes_out"`
		}{
			BytesIn:  int64(1024000 + rand.Intn(512000)),
			BytesOut: int64(2048000 + rand.Intn(1024000)),
		},
		ActiveConnections: 45 + rand.Intn(55),
		RequestsPerMinute: 1200 + rand.Intn(300),
	}

	// Cache the result
	go func() {
		cacheCtx := context.Background()
		if err := s.redis.CacheSystemMetrics(cacheCtx, metrics, 30*time.Second); err != nil {
			log.Printf("Failed to cache system metrics: %v", err)
		} else {
			log.Printf("CACHE WRITE: GetSystemMetrics")
		}
	}()

	return metrics, nil
}

func (s *MonitoringService) GetUptime(ctx context.Context) (*models.UptimeData, error) {
	// Try to get from cache first
	if cachedUptime, err := s.redis.GetCachedUptime(ctx); err == nil && cachedUptime != nil {
		log.Printf("CACHE HIT: GetUptime")
		return cachedUptime, nil
	}

	log.Printf("CACHE MISS: GetUptime - generating uptime data")

	// For demo purposes, generate mock uptime data
	// In a real implementation, this would track actual uptime
	uptime := &models.UptimeData{
		CurrentUptimeHours: 720.5 + rand.Float64()*100,
		UptimePercent24h:   100.0,
		UptimePercent7d:    99.8 + rand.Float64()*0.2,
		UptimePercent30d:   99.95 + rand.Float64()*0.05,
		IncidentsCount:     rand.Intn(5),
		LastDowntime:       nil, // No recent downtime
	}

	// Randomly add a last downtime
	if rand.Float64() < 0.3 {
		lastDowntime := time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour)
		uptime.LastDowntime = &lastDowntime
	}

	// Cache the result
	go func() {
		cacheCtx := context.Background()
		if err := s.redis.CacheUptime(cacheCtx, uptime, 5*time.Minute); err != nil {
			log.Printf("Failed to cache uptime: %v", err)
		} else {
			log.Printf("CACHE WRITE: GetUptime")
		}
	}()

	return uptime, nil
}
