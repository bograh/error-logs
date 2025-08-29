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
	ErrorQueueKey              = "error_queue"
	RecentErrorsKey            = "recent_errors"
	ErrorCachePrefix           = "error_cache:"
	StatsCacheKey              = "stats_cache"
	TrendsCachePrefix          = "trends_cache:"
	PerformanceMetricsCacheKey = "performance_metrics_cache"
	ServiceHealthCacheKey      = "service_health_cache"
	SystemMetricsCacheKey      = "system_metrics_cache"
	UptimeCacheKey             = "uptime_cache"
	CacheKeysSetKey            = "cache_keys_set"
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

// Analytics caching methods
func (c *Client) CacheTrends(ctx context.Context, key string, trends *models.TrendResponse, ttl time.Duration) error {
	start := time.Now()

	trendsJSON, err := json.Marshal(trends)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: Trends - key: %s, error: %v", key, err)
		return fmt.Errorf("failed to marshal trends: %w", err)
	}

	fullKey := TrendsCachePrefix + key
	err = c.Set(ctx, fullKey, trendsJSON, ttl).Err()
	if err != nil {
		log.Printf("REDIS WRITE ERROR: Trends - key: %s, error: %v, duration: %v", key, err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: Trends - key: %s, ttl: %v, duration: %v", key, ttl, time.Since(start))
	return nil
}

func (c *Client) GetCachedTrends(ctx context.Context, key string) (*models.TrendResponse, error) {
	start := time.Now()
	fullKey := TrendsCachePrefix + key

	result, err := c.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: Trends - key: %s, duration: %v", key, time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedTrends - key: %s, error: %v, duration: %v", key, err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached trends: %w", err)
	}

	var trends models.TrendResponse
	if err := json.Unmarshal([]byte(result), &trends); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: Trends - key: %s, error: %v", key, err)
		return nil, fmt.Errorf("failed to unmarshal cached trends: %w", err)
	}

	log.Printf("REDIS CACHE HIT: Trends - key: %s, duration: %v", key, time.Since(start))
	return &trends, nil
}

func (c *Client) CachePerformanceMetrics(ctx context.Context, key string, metrics *models.PerformanceMetrics, ttl time.Duration) error {
	start := time.Now()

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: Performance metrics - key: %s, error: %v", key, err)
		return fmt.Errorf("failed to marshal performance metrics: %w", err)
	}

	err = c.Set(ctx, key, metricsJSON, ttl).Err()
	if err != nil {
		log.Printf("REDIS WRITE ERROR: Performance metrics - key: %s, error: %v, duration: %v", key, err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: Performance metrics - key: %s, ttl: %v, duration: %v", key, ttl, time.Since(start))
	return nil
}

func (c *Client) GetCachedPerformanceMetrics(ctx context.Context, key string) (*models.PerformanceMetrics, error) {
	start := time.Now()

	result, err := c.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: Performance metrics - key: %s, duration: %v", key, time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedPerformanceMetrics - key: %s, error: %v, duration: %v", key, err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached performance metrics: %w", err)
	}

	var metrics models.PerformanceMetrics
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: Performance metrics - key: %s, error: %v", key, err)
		return nil, fmt.Errorf("failed to unmarshal cached performance metrics: %w", err)
	}

	log.Printf("REDIS CACHE HIT: Performance metrics - key: %s, duration: %v", key, time.Since(start))
	return &metrics, nil
}

// Monitoring caching methods
func (c *Client) CacheServiceHealth(ctx context.Context, services *models.ServicesResponse, ttl time.Duration) error {
	start := time.Now()

	servicesJSON, err := json.Marshal(services)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: Service health - error: %v", err)
		return fmt.Errorf("failed to marshal service health: %w", err)
	}

	err = c.Set(ctx, ServiceHealthCacheKey, servicesJSON, ttl).Err()
	if err != nil {
		log.Printf("REDIS WRITE ERROR: Service health - error: %v, duration: %v", err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: Service health - ttl: %v, duration: %v", ttl, time.Since(start))
	return nil
}

func (c *Client) GetCachedServiceHealth(ctx context.Context) (*models.ServicesResponse, error) {
	start := time.Now()

	result, err := c.Get(ctx, ServiceHealthCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: Service health - duration: %v", time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedServiceHealth - error: %v, duration: %v", err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached service health: %w", err)
	}

	var services models.ServicesResponse
	if err := json.Unmarshal([]byte(result), &services); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: Service health - error: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cached service health: %w", err)
	}

	log.Printf("REDIS CACHE HIT: Service health - duration: %v", time.Since(start))
	return &services, nil
}

func (c *Client) CacheSystemMetrics(ctx context.Context, metrics *models.SystemMetrics, ttl time.Duration) error {
	start := time.Now()

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: System metrics - error: %v", err)
		return fmt.Errorf("failed to marshal system metrics: %w", err)
	}

	err = c.Set(ctx, SystemMetricsCacheKey, metricsJSON, ttl).Err()
	if err != nil {
		log.Printf("REDIS WRITE ERROR: System metrics - error: %v, duration: %v", err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: System metrics - ttl: %v, duration: %v", ttl, time.Since(start))
	return nil
}

func (c *Client) GetCachedSystemMetrics(ctx context.Context) (*models.SystemMetrics, error) {
	start := time.Now()

	result, err := c.Get(ctx, SystemMetricsCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: System metrics - duration: %v", time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedSystemMetrics - error: %v, duration: %v", err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached system metrics: %w", err)
	}

	var metrics models.SystemMetrics
	if err := json.Unmarshal([]byte(result), &metrics); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: System metrics - error: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cached system metrics: %w", err)
	}

	log.Printf("REDIS CACHE HIT: System metrics - duration: %v", time.Since(start))
	return &metrics, nil
}

func (c *Client) CacheUptime(ctx context.Context, uptime *models.UptimeData, ttl time.Duration) error {
	start := time.Now()

	uptimeJSON, err := json.Marshal(uptime)
	if err != nil {
		log.Printf("REDIS MARSHAL ERROR: Uptime - error: %v", err)
		return fmt.Errorf("failed to marshal uptime: %w", err)
	}

	err = c.Set(ctx, UptimeCacheKey, uptimeJSON, ttl).Err()
	if err != nil {
		log.Printf("REDIS WRITE ERROR: Uptime - error: %v, duration: %v", err, time.Since(start))
		return err
	}

	log.Printf("REDIS CACHE WRITE: Uptime - ttl: %v, duration: %v", ttl, time.Since(start))
	return nil
}

func (c *Client) GetCachedUptime(ctx context.Context) (*models.UptimeData, error) {
	start := time.Now()

	result, err := c.Get(ctx, UptimeCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("REDIS CACHE MISS: Uptime - duration: %v", time.Since(start))
			return nil, nil
		}
		log.Printf("REDIS ERROR: GetCachedUptime - error: %v, duration: %v", err, time.Since(start))
		return nil, fmt.Errorf("failed to get cached uptime: %w", err)
	}

	var uptime models.UptimeData
	if err := json.Unmarshal([]byte(result), &uptime); err != nil {
		log.Printf("REDIS UNMARSHAL ERROR: Uptime - error: %v", err)
		return nil, fmt.Errorf("failed to unmarshal cached uptime: %w", err)
	}

	log.Printf("REDIS CACHE HIT: Uptime - duration: %v", time.Since(start))
	return &uptime, nil
}
