package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"error-logs/internal/database"
	"error-logs/internal/models"
	"error-logs/internal/redis"
)

type ErrorService struct {
	db    *database.DB
	redis *redis.Client
}

func NewErrorService(db *database.DB, redis *redis.Client) *ErrorService {
	return &ErrorService{
		db:    db,
		redis: redis,
	}
}

func (s *ErrorService) CreateError(ctx context.Context, req *models.CreateErrorRequest, userAgent, ipAddress string) (*models.Error, error) {
	now := time.Now().UTC()

	// Generate fingerprint for grouping similar errors
	fingerprint := generateFingerprint(req.Message, req.StackTrace)

	error := &models.Error{
		ID:          uuid.New(),
		Timestamp:   now,
		Level:       req.Level,
		Message:     req.Message,
		StackTrace:  req.StackTrace,
		Context:     req.Context,
		Source:      req.Source,
		Environment: "production", // Default
		UserAgent:   &userAgent,
		IPAddress:   &ipAddress,
		URL:         req.URL,
		Fingerprint: &fingerprint,
		Resolved:    false,
		Count:       1,
		FirstSeen:   now,
		LastSeen:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.Environment != nil {
		error.Environment = *req.Environment
	}

	if error.Context == nil {
		error.Context = make(map[string]interface{})
	}

	// Queue error for processing
	if err := s.redis.QueueError(ctx, error); err != nil {
		log.Printf("Failed to queue error to Redis: %v", err)
		// Fall back to direct database insert
		if err := s.db.CreateError(error); err != nil {
			return nil, err
		}
		// Clear cache after successful database insert
		if err := s.redis.InvalidateAllCache(ctx); err != nil {
			log.Printf("Failed to invalidate cache after direct insert: %v", err)
		}
		return error, nil
	}

	// Invalidate cache
	if err := s.redis.InvalidateAllCache(ctx); err != nil {
		log.Printf("Failed to invalidate cache after queuing error: %v", err)
	}

	return error, nil
}

func (s *ErrorService) GetErrors(ctx context.Context, limit, offset int, level, source string) (*models.ErrorListResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("list_%d_%d_%s_%s", limit, offset, level, source)
	if cachedErrors, err := s.redis.GetCachedErrorList(ctx, cacheKey); err == nil && cachedErrors != nil {
		total := len(cachedErrors) + offset // Approximate
		return &models.ErrorListResponse{
			Errors: cachedErrors,
			Total:  total,
			Page:   (offset / limit) + 1,
			Limit:  limit,
		}, nil
	}

	// Get from database
	errors, total, err := s.db.GetErrors(limit, offset, level, source)
	if err != nil {
		return nil, err
	}

	// Cache results
	if len(errors) > 0 {
		s.redis.CacheErrorList(ctx, cacheKey, errors, 2*time.Minute)
	}

	return &models.ErrorListResponse{
		Errors: errors,
		Total:  total,
		Page:   (offset / limit) + 1,
		Limit:  limit,
	}, nil
}

func (s *ErrorService) GetErrorByID(ctx context.Context, id uuid.UUID) (*models.Error, error) {
	return s.db.GetErrorByID(id)
}

func (s *ErrorService) ResolveError(ctx context.Context, id uuid.UUID) error {
	err := s.db.ResolveError(id)
	if err != nil {
		return err
	}

	// Invalidate both error lists and stats cache since resolving changes stats
	if err := s.redis.InvalidateAllCache(ctx); err != nil {
		log.Printf("Failed to invalidate cache after resolving error: %v", err)
	}
	return nil
}

func (s *ErrorService) DeleteError(ctx context.Context, id uuid.UUID) error {
	err := s.db.DeleteError(id)
	if err != nil {
		return err
	}

	if err := s.redis.InvalidateAllCache(ctx); err != nil {
		log.Printf("Failed to invalidate cache after deleting error: %v", err)
	}
	return nil
}

func (s *ErrorService) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	// Try cache first
	if cachedStats, err := s.redis.GetCachedStats(ctx); err == nil && cachedStats != nil {
		log.Printf("Returning cached stats: %+v", cachedStats)
		return cachedStats, nil
	}

	// Get from database
	stats, err := s.db.GetStats()
	if err != nil {
		log.Printf("Failed to get stats from database: %v", err)
		return nil, err
	}

	log.Printf("Retrieved stats from database: %+v", stats)

	if err := s.redis.CacheStats(ctx, stats); err != nil {
		log.Printf("Failed to cache stats: %v", err)
	}

	return stats, nil
}

func (s *ErrorService) StartQueueProcessor(ctx context.Context) {
	log.Println("Starting error queue processor...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Queue processor stopped")
			return
		default:
			error, err := s.redis.DequeueError(ctx)
			if err != nil {
				log.Printf("Failed to dequeue error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if error == nil {
				continue // No error available
			}

			if err := s.processError(ctx, error); err != nil {
				log.Printf("Failed to process error: %v", err)
			}
		}
	}
}

func (s *ErrorService) processError(ctx context.Context, error *models.Error) error {
	if error.Fingerprint != nil {
	}

	err := s.db.CreateError(error)
	if err != nil {
		return err
	}

	if err := s.redis.InvalidateAllCache(ctx); err != nil {
		log.Printf("Failed to invalidate cache after processing error: %v", err)
	}

	return nil
}

func generateFingerprint(message string, stackTrace *string) string {
	data := message
	if stackTrace != nil {
		data += *stackTrace
	}

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 characters
}
