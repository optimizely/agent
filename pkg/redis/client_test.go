package redis

import (
	"fmt"
	"testing"

	"github.com/optimizely/sidedoor/config"
)

func TestNewClient(t *testing.T) {
	conf := config.RedisConfig{
		Address:  "localhost:6379",
		Password: "", // no password set
		Database: 0,  // use default DB
	}

	client := NewClient(conf)

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}
