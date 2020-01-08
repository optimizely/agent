package redis

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/optimizely/sidedoor/config"
)

func NewClient(conf config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Address,
		Password: conf.Password,
		DB:       conf.Database,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>

	return client
}
