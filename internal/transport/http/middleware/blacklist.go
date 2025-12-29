package middleware

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBlacklist struct {
	rdb    *redis.Client
	prefix string
}

func NewRedisTokenBlacklist(rdb *redis.Client, prefix string) *RedisTokenBlacklist {
	if prefix == "" {
		prefix = "revoked:jti:"
	}
	return &RedisTokenBlacklist{rdb: rdb, prefix: prefix}
}

func (b *RedisTokenBlacklist) key(jti string) string { return b.prefix + jti }

func (b *RedisTokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := b.rdb.Exists(ctx, b.key(jti)).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (b *RedisTokenBlacklist) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Minute
	}
	return b.rdb.Set(ctx, b.key(jti), "1", ttl).Err()
}
