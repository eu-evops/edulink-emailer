package common

import (
	"context"
	"time"
)

type CacheInt interface {
	Initialise() error
	Get(ctx context.Context, key string, value interface{}) error
	Set(item *Item) error
	Exists(ctx context.Context, key string) bool
}

type CacheType int

const (
	Redis CacheType = iota
	Local
)

type CacheOptions struct {
	CacheType     CacheType
	RedisHost     string
	RedisUsername string
	RedisPassword string
}

type Item struct {
	Ctx   context.Context
	Key   string
	Value interface{}
	TTL   time.Duration
}
