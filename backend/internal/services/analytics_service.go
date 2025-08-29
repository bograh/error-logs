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

type AnalyticsService struct {
	db    *database.DB
	redis *redis.Client
}

func NewAnalyticsService(db *database.DB, redis *redis.Client) *AnalyticsService {
	return &AnalyticsService{
		db:    db,
		redis: redis,
	}
}

func (s *AnalyticsService) GetTrends(ctx context.Context, period, groupBy string) (*models.TrendResponse, error) {
	cacheKey := "trends_" + period + "_" + groupBy

	// Try to get from cache first
	if cachedTrends, err := s.redis.GetCachedTrends(ctx, cacheKey); err == nil && cachedTrends != nil {
		log.Printf("CACHE HIT: GetTrends - key: %s", cacheKey)
		return cachedTrends, nil
	}

	log.Printf("CACHE MISS: GetTrends - key: %s, fetching from database", cacheKey)

	trends, err := s.db.GetTrends(period, groupBy)
	if err != nil {
		return nil, err
	}

	// Cache the result
	go func() {
		cacheCtx := context.Background()
		if err := s.redis.CacheTrends(cacheCtx, cacheKey, trends, 5*time.Minute); err != nil {
			log.Printf("Failed to cache trends: %v", err)
		} else {
			log.Printf("CACHE WRITE: GetTrends - key: %s", cacheKey)
		}
	}()

	return trends, nil
}

func (s *AnalyticsService) GetPerformanceMetrics(ctx context.Context) (*models.PerformanceMetrics, error) {
	cacheKey := "performance_metrics"

	// Try to get from cache first
	if cachedMetrics, err := s.redis.GetCachedPerformanceMetrics(ctx, cacheKey); err == nil && cachedMetrics != nil {
		log.Printf("CACHE HIT: GetPerformanceMetrics - key: %s", cacheKey)
		return cachedMetrics, nil
	}

	log.Printf("CACHE MISS: GetPerformanceMetrics - key: %s, generating mock data", cacheKey)

	// For demo purposes, generate mock performance metrics
	// In a real implementation, this would collect actual system metrics
	metrics := &models.PerformanceMetrics{
		AvgResponseTime:     245 + rand.Intn(100),
		ErrorRatePercent:    0.8 + rand.Float64()*0.5,
		ThroughputRPM:       1200 + rand.Intn(200),
		AvailabilityPercent: 99.95 - rand.Float64()*0.1,
		PerformanceScore:    8.7 + rand.Float64()*0.6,
	}

	// Cache the result
	go func() {
		cacheCtx := context.Background()
		if err := s.redis.CachePerformanceMetrics(cacheCtx, cacheKey, metrics, 1*time.Minute); err != nil {
			log.Printf("Failed to cache performance metrics: %v", err)
		} else {
			log.Printf("CACHE WRITE: GetPerformanceMetrics - key: %s", cacheKey)
		}
	}()

	return metrics, nil
}
