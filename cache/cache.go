package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"miikka.xyz/devops-app/lib/repo"
	"miikka.xyz/devops-app/utils"
)

const redisKeyTraffic = "traffic"

// Cache is wrapper for a redis client
type Cache struct {
	*redis.Client
}

// New returns a new redis client
func New(isTest bool) (*Cache, func(ctx context.Context)) {
	db := 0
	if isTest {
		db = 3
	}

	cache := &Cache{
		redis.NewClient(&redis.Options{
			Addr:     utils.GetEnv("REDIS_URL", "localhost:6379"),
			Password: "",
			DB:       db,
		}),
	}

	teardown := func(ctx context.Context) {
		cache.FlushDB(ctx)
	}

	return cache, teardown
}

// UpdateTrafficCache gets traffic data from database starting from day 'daysAgo'
// and puts that data to cache. Default 'daysAgo' is 8
func (c *Cache) UpdateTrafficCache(ctx context.Context, daysAgo int) error {
	log.Println("updating cache...")

	// Default is one week. -8 days because GitHub API doesn't return data for current day (not sure though)
	if daysAgo == 0 {
		daysAgo = -8
	} else if daysAgo >= 0 {
		daysAgo = -daysAgo
	}

	// Get traffic data from database
	repos, err := repo.StoreGetRepositoryTrafficWithMeta(ctx, time.Now().AddDate(0, 0, daysAgo))
	if err != nil {
		return err
	}

	// Clear old data from cache
	status := c.FlushDB(ctx)
	if status.Err() != nil {
		return err
	}

	// Format that data and update cache
	if err := c.updateTrafficData(ctx, repo.FormatRepositorysToMap(repos)); err != nil {
		log.Println("updating traffic cache failed", err)
		return err
	}

	log.Println("cache updated!")
	return nil
}

// GetTrafficData returns traffic data from cache
func (c *Cache) GetTrafficData(ctx context.Context) (repo.ReposByNameMap, error) {
	cacheData, err := c.Get(ctx, redisKeyTraffic).Result()
	if err != nil {
		return nil, err
	}

	var trafficByRepoName repo.ReposByNameMap
	err = json.Unmarshal([]byte(cacheData), &trafficByRepoName)
	if err != nil {
		return nil, err
	}

	return trafficByRepoName, nil
}

func (c *Cache) updateTrafficData(ctx context.Context, trafficByRepoName repo.ReposByNameMap) error {
	err := c.Set(ctx, redisKeyTraffic, trafficByRepoName, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
