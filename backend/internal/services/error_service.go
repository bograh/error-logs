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
	fingerprint := generateFingerprint(req.Message, req.StackTrace)

	error := &models.Error{
		ID:          uuid.New(),
		Timestamp:   now,
		Level:       req.Level,
		Message:     req.Message,
		StackTrace:  req.StackTrace,
		Context:     req.Context,
		Source:      req.Source,
		Environment: "production",
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

	if err := s.redis.QueueError(ctx, error); err != nil {
		log.Printf("Failed to queue error to Redis: %v", err)
		if err := s.db.CreateError(error); err != nil {
			return nil, err
		}
		log.Printf("CACHE INVALIDATION: CreateError (fallback) - invalidating all caches")
		s.redis.InvalidateAllCache(context.Background())
		return error, nil
	}

	log.Printf("CACHE INVALIDATION: CreateError - invalidating all caches")
	s.redis.InvalidateAllCache(context.Background())
	return error, nil
}

func (s *ErrorService) GetErrors(ctx context.Context, limit, offset int, level, source string) (*models.ErrorListResponse, error) {
	cacheKey := fmt.Sprintf("list_%d_%d_%s_%s", limit, offset, level, source)
	start := time.Now()

	if cachedErrors, err := s.redis.GetCachedErrorList(ctx, cacheKey); err == nil && cachedErrors != nil {
		log.Printf("CACHE HIT: GetErrors - key: %s, duration: %v", cacheKey, time.Since(start))
		return &models.ErrorListResponse{
			Errors: cachedErrors,
			Total:  len(cachedErrors) + offset,
			Page:   (offset / limit) + 1,
			Limit:  limit,
		}, nil
	}

	log.Printf("CACHE MISS: GetErrors - key: %s, fetching from database", cacheKey)
	errors, total, err := s.db.GetErrors(limit, offset, level, source)
	if err != nil {
		return nil, err
	}

	dbDuration := time.Since(start)
	log.Printf("DATABASE QUERY: GetErrors completed in %v", dbDuration)

	if len(errors) > 0 {
		go func() {
			cacheStart := time.Now()
			// Use background context to avoid cancellation when HTTP request ends
			cacheCtx := context.Background()
			if err := s.redis.CacheErrorList(cacheCtx, cacheKey, errors, 2*time.Minute); err != nil {
				log.Printf("Failed to cache error list: %v", err)
			} else {
				log.Printf("CACHE WRITE: GetErrors - key: %s, duration: %v", cacheKey, time.Since(cacheStart))
			}
		}()
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
	if err := s.db.ResolveError(id); err != nil {
		return err
	}
	log.Printf("CACHE INVALIDATION: ResolveError - invalidating all caches for error ID: %s", id)
	go s.redis.InvalidateAllCache(context.Background())
	return nil
}

func (s *ErrorService) DeleteError(ctx context.Context, id uuid.UUID) error {
	if err := s.db.DeleteError(id); err != nil {
		return err
	}
	log.Printf("CACHE INVALIDATION: DeleteError - invalidating all caches for error ID: %s", id)
	go s.redis.InvalidateAllCache(context.Background())
	return nil
}

func (s *ErrorService) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	start := time.Now()

	if cachedStats, err := s.redis.GetCachedStats(ctx); err == nil && cachedStats != nil {
		log.Printf("CACHE HIT: GetStats - duration: %v", time.Since(start))
		return cachedStats, nil
	}

	log.Printf("CACHE MISS: GetStats - fetching from database")
	stats, err := s.db.GetStats()
	if err != nil {
		return nil, err
	}

	dbDuration := time.Since(start)
	log.Printf("DATABASE QUERY: GetStats completed in %v", dbDuration)

	go func() {
		cacheStart := time.Now()
		// Use background context to avoid cancellation when HTTP request ends
		cacheCtx := context.Background()
		if err := s.redis.CacheStats(cacheCtx, stats); err != nil {
			log.Printf("Failed to cache stats: %v", err)
		} else {
			log.Printf("CACHE WRITE: GetStats - duration: %v", time.Since(cacheStart))
		}
	}()

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
	if err := s.db.CreateError(error); err != nil {
		return err
	}
	log.Printf("CACHE INVALIDATION: processError - invalidating all caches for processed error")
	go s.redis.InvalidateAllCache(context.Background())
	return nil
}

func generateFingerprint(message string, stackTrace *string) string {
	data := message
	if stackTrace != nil {
		data += *stackTrace
	}
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}
