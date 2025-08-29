package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	CacheKeysSetKey  = "cache_keys_set"
)

func (c *Client) QueueError(ctx context.Context, error *models.Error) error {
	errorJSON, err := json.Marshal(error)
	if err != nil {
		return fmt.Errorf("failed to marshal error: %w", err)
	}

	pipe := c.Pipeline()
	pipe.LPush(ctx, ErrorQueueKey, errorJSON)
	pipe.LPush(ctx, RecentErrorsKey, errorJSON)
	pipe.LTrim(ctx, RecentErrorsKey, 0, 99)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *Client) DequeueError(ctx context.Context) (*models.Error, error) {
	result, err := c.BRPop(ctx, 5*time.Second, ErrorQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue error: %w", err)
	}

	var error models.Error
	if err := json.Unmarshal([]byte(result[1]), &error); err != nil {
		return nil, fmt.Errorf("failed to unmarshal error: %w", err)
	}
	return &error, nil
}

func (c *Client) GetRecentErrors(ctx context.Context, limit int) ([]models.Error, error) {
	results, err := c.LRange(ctx, RecentErrorsKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent errors: %w", err)
	}

	errors := make([]models.Error, 0, len(results))
	for _, result := range results {
		var error models.Error
		if err := json.Unmarshal([]byte(result), &error); err != nil {
			continue
		}
		errors = append(errors, error)
	}
	return errors, nil
}

func (c *Client) CacheErrorList(ctx context.Context, key string, errors []models.Error, ttl time.Duration) error {
	start := time.Now()

	errorsJSON, err := json.Marshal(errors)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: Error list - key: %s, error: %v", key, err)
		return fmt.Errorf("failed to marshal errors: %w", err)
	}

	fullKey := ErrorCachePrefix + key
	pipe := c.Pipeline()
	pipe.Set(ctx, fullKey, errorsJSON, ttl)
	pipe.SAdd(ctx, CacheKeysSetKey, fullKey)
	_, err = pipe.Exec(ctx)

	if err != nil {
		log.Printf("REDIS WRITE ERROR: Error list - key: %s, error: %v, duration: %v", key, err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: Error list - key: %s, count: %d, ttl: %v, duration: %v", key, len(errors), ttl, time.Since(start))
	return nil
}

func (c *Client) GetCachedErrorList(ctx context.Context, key string) ([]models.Error, error) {
	start := time.Now()
	fullKey := ErrorCachePrefix + key

	result, err := c.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: Error list - key: %s, duration: %v", key, time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedErrorList - key: %s, error: %v, duration: %v", key, err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached errors: %w", err)
	}

	var errors []models.Error
	if err := json.Unmarshal([]byte(result), &errors); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: Error list - key: %s, error: %v", key, err)
		return nil, fmt.Errorf("failed to unmarshal cached errors: %w", err)
	}

	log.Printf("REDIS CACHE HIT: Error list - key: %s, count: %d, duration: %v", key, len(errors), time.Since(start))
	return errors, nil
}

func (c *Client) CacheStats(ctx context.Context, stats *models.StatsResponse) error {
	start := time.Now()

	statsJSON, err := json.Marshal(stats)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: Stats - error: %v", err)
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	err = c.Set(ctx, StatsCacheKey, statsJSON, 5*time.Minute).Err()
	if err != nil {
		log.Printf("REDIS WRITE ERROR: Stats - error: %v, duration: %v", err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: Stats - ttl: 5m, duration: %v", time.Since(start))
	return nil
}

func (c *Client) GetCachedStats(ctx context.Context) (*models.StatsResponse, error) {
	start := time.Now()

	result, err := c.Get(ctx, StatsCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: Stats - duration: %v", time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedStats - error: %v, duration: %v", err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached stats: %w", err)
	}

	var stats models.StatsResponse
	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: Stats - error: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cached stats: %w", err)
	}

	log.Printf("REDIS CACHE HIT: Stats - duration: %v", time.Since(start))
	return &stats, nil
}

func (c *Client) InvalidateErrorCache(ctx context.Context) error {
	start := time.Now()

	keys, err := c.Keys(ctx, ErrorCachePrefix+"*").Result()
	if err != nil {
		log.Printf("REDIS INVALIDATE ERROR: Error cache - failed to get keys: %v", err)
		return err
	}

	if len(keys) > 0 {
		err = c.Del(ctx, keys...).Err()
		if err != nil {
			log.Printf("REDIS INVALIDATE ERROR: Error cache - failed to delete keys: %v", err)
			return err
		}
		log.Printf("REDIS CACHE INVALIDATE: Error cache - deleted %d keys, duration: %v", len(keys), time.Since(start))
	} else {
		log.Printf("REDIS CACHE INVALIDATE: Error cache - no keys to delete, duration: %v", time.Since(start))
	}

	return nil
}

func (c *Client) InvalidateStatsCache(ctx context.Context) error {
	start := time.Now()

	err := c.Del(ctx, StatsCacheKey).Err()
	if err != nil {
		log.Printf("REDIS INVALIDATE ERROR: Stats cache - error: %v", err)
		return err
	}

	log.Printf("REDIS CACHE INVALIDATE: Stats cache - duration: %v", time.Since(start))
	return nil
}

func (c *Client) InvalidateAllCache(ctx context.Context) error {
	start := time.Now()
	log.Printf("REDIS CACHE INVALIDATE: Starting full cache invalidation")

	if err := c.InvalidateErrorCache(ctx); err != nil {
		return err
	}

	err := c.InvalidateStatsCache(ctx)
	log.Printf("REDIS CACHE INVALIDATE: Full cache invalidation completed, duration: %v", time.Since(start))
	return err
}
