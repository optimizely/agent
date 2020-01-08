package redis

import (
	"github.com/go-redis/redis/v7"

	"github.com/optimizely/sidedoor/config"
)

func NewClient(conf config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Address,
		Password: conf.Password,
		DB:       conf.Database,
	})

	return client
}
