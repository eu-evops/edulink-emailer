package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eu-evops/edulink/pkg/cache/common"
	cachev9 "github.com/go-redis/cache/v9"
	"github.com/go-redis/redis/v9"
)

type RedisCache struct {
	cache *cachev9.Cache

	options *common.CacheOptions
}

func New(options *common.CacheOptions) *RedisCache {
	return &RedisCache{
		options: options,
	}
}

func (c *RedisCache) Initialise() error {
	log.Println("Initialising cache")
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", c.options.RedisHost),
		Username: c.options.RedisUsername,
		Password: c.options.RedisPassword,
	})
	c.cache = cachev9.New(&cachev9.Options{
		Redis:        redis,
		LocalCache:   cachev9.NewTinyLFU(1000, time.Minute),
		StatsEnabled: true,
	})

	var ret string
	err := c.Get(context.Background(), "randomKey", &ret)
	if err != nil && err != cachev9.ErrCacheMiss {
		return err
	}

	return nil
}

func (cd *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
	log.Printf("Getting key %s from cache %+v", key, cd)
	return cd.cache.Get(ctx, key, value)
}

func cachev9Version(i *common.Item) *cachev9.Item {
	return &cachev9.Item{
		Key:   i.Key,
		Value: i.Value,
		TTL:   i.TTL,
		Ctx:   i.Ctx,
	}
}

func (c *RedisCache) Set(item *common.Item) error {
	return c.cache.Set(cachev9Version(item))
}

func (c *RedisCache) Exists(ctx context.Context, key string) bool {
	return c.cache.Exists(ctx, key)
}
