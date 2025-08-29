package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"error-logs/internal/models"
)

type Client struct {
	*redis.Client
}

func NewClient(redisURL string) (*Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb}, nil
}

const (
	ErrorQueueKey    = "error_queue"
	RecentErrorsKey  = "recent_errors"
	ErrorCachePrefix = "error_cache:"
	StatsCacheKey    = "stats_cache"
)

// QueueError adds an error to the processing queue
func (c *Client) QueueError(ctx context.Context, error *models.Error) error {
	errorJSON, err := json.Marshal(error)
	if err != nil {
		return fmt.Errorf("failed to marshal error: %w", err)
	}

	// Add to processing queue
	err = c.LPush(ctx, ErrorQueueKey, errorJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to queue error: %w", err)
	}

	// Add to recent errors list (keep last 100)
	pipe := c.Pipeline()
	pipe.LPush(ctx, RecentErrorsKey, errorJSON)
	pipe.LTrim(ctx, RecentErrorsKey, 0, 99) // Keep only last 100 errors
	_, err = pipe.Exec(ctx)

	return err
}

// DequeueError removes and returns an error from the processing queue
func (c *Client) DequeueError(ctx context.Context) (*models.Error, error) {
	result, err := c.BRPop(ctx, 5*time.Second, ErrorQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No error available
		}
		return nil, fmt.Errorf("failed to dequeue error: %w", err)
	}

	var error models.Error
	if err := json.Unmarshal([]byte(result[1]), &error); err != nil {
		return nil, fmt.Errorf("failed to unmarshal error: %w", err)
	}

	return &error, nil
}

// GetRecentErrors retrieves recent errors from cache
func (c *Client) GetRecentErrors(ctx context.Context, limit int) ([]models.Error, error) {
	results, err := c.LRange(ctx, RecentErrorsKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent errors: %w", err)
	}

	var errors []models.Error
	for _, result := range results {
		var error models.Error
		if err := json.Unmarshal([]byte(result), &error); err != nil {
			continue // Skip malformed entries
		}
		errors = append(errors, error)
	}

	return errors, nil
}

// CacheErrorList caches a list of errors with TTL
func (c *Client) CacheErrorList(ctx context.Context, key string, errors []models.Error, ttl time.Duration) error {
	errorsJSON, err := json.Marshal(errors)
	if err != nil {
		return fmt.Errorf("failed to marshal errors: %w", err)
	}

	return c.Set(ctx, ErrorCachePrefix+key, errorsJSON, ttl).Err()
}

// GetCachedErrorList retrieves cached error list
func (c *Client) GetCachedErrorList(ctx context.Context, key string) ([]models.Error, error) {
	result, err := c.Get(ctx, ErrorCachePrefix+key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached errors: %w", err)
	}

	var errors []models.Error
	if err := json.Unmarshal([]byte(result), &errors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached errors: %w", err)
	}

	return errors, nil
}

// CacheStats caches statistics with TTL
func (c *Client) CacheStats(ctx context.Context, stats *models.StatsResponse) error {
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return c.Set(ctx, StatsCacheKey, statsJSON, 5*time.Minute).Err()
}

// GetCachedStats retrieves cached statistics
func (c *Client) GetCachedStats(ctx context.Context) (*models.StatsResponse, error) {
	result, err := c.Get(ctx, StatsCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached stats: %w", err)
	}

	var stats models.StatsResponse
	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached stats: %w", err)
	}

	return &stats, nil
}

func (c *Client) InvalidateErrorCache(ctx context.Context) error {
	keys, err := c.Keys(ctx, ErrorCachePrefix+"*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.Del(ctx, keys...).Err()
	}

	return nil
}

func (c *Client) InvalidateStatsCache(ctx context.Context) error {
	return c.Del(ctx, StatsCacheKey).Err()
}

func (c *Client) InvalidateAllCache(ctx context.Context) error {
	// Clear error lists cache
	if err := c.InvalidateErrorCache(ctx); err != nil {
		return err
	}

	// Clear stats cache
	return c.InvalidateStatsCache(ctx)
}
