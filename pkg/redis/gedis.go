package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Gedis struct {
	clientredis *redis.Client
}

type GedisConfig struct {
	Addess   string
	Password string
	Username string
}

func NewRedisClient(c GedisConfig) *Gedis {
	return &Gedis{
		clientredis: redis.NewClient(
			&redis.Options{
				Addr:     c.Addess,
				Password: c.Password,
				Username: c.Username,
			},
		),
	}
}

func (g *Gedis) Get(key string) *redis.MapStringStringCmd {
	return g.clientredis.HGetAll(context.Background(), key)

}

func (g *Gedis) Set(key string, value map[string]string) *redis.IntCmd {
	result := g.clientredis.HSet(context.Background(), key, value)

	if result.Err() != redis.Nil {
		g.clientredis.Expire(context.Background(), key, time.Hour)
	}
	return result
}

func (g *Gedis) IsExistsKey(key string) *redis.IntCmd {
	return g.clientredis.Exists(context.Background(), key)
}

func (g *Gedis) IsExistsKeyField(key string, field string) *redis.BoolCmd {
	return g.clientredis.HExists(context.Background(), key, field)
}

func (g *Gedis) DelKeyWithPattern(pattern string) *redis.IntCmd {
	return g.clientredis.Del(context.Background(),
		g.clientredis.Keys(context.Background(), pattern).Val()...)
}

func (g *Gedis) GetAllKeyWithPattern(pattern string) []string {
	return g.clientredis.Keys(context.Background(), pattern).Val()
}
