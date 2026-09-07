// Package cache adapts Redis for Auth.
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

var incrementWithTTL = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

func New(addr string) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // "localhost:6379"
	})
	return &Cache{client: client}
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}

	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *Cache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return incrementWithTTL.Run(ctx, c.client, []string{key}, ttl.Milliseconds()).Int64()
}
