package cache

import (
	"context"
	"log"

	"github.com/eu-evops/edulink/pkg/cache/common"
	"github.com/eu-evops/edulink/pkg/cache/redis"
)

type Cache struct {
	cache common.CacheInt
}

func New(options *common.CacheOptions) *Cache {
	c := &Cache{}

	switch options.CacheType {
	case common.Redis:
		c.cache = redis.New(options)
	default:
		return nil
	}

	return c
}

func (c *Cache) Initialise() error {
	log.Println("Initialising cache")
	return c.cache.Initialise()
}

func (c *Cache) Get(ctx context.Context, key string, value interface{}) error {
	return c.cache.Get(ctx, key, value)
}

func (c *Cache) Set(item *common.Item) error {
	return c.cache.Set(item)
}

func (c *Cache) Exists(ctx context.Context, key string) bool {
	return c.cache.Exists(ctx, key)
}
