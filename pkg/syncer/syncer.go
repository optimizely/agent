package syncer

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/go-sdk/pkg/notification"
)

type CustomSyncer struct {
}

func (s *CustomSyncer) AddHandler(_ notification.Type, _ func(interface{})) (int, error) {
	return 0, nil
}

func (s *CustomSyncer) RemoveHandler(_ int, _ notification.Type) error {
	return nil
}

func (s *CustomSyncer) Send(notificationType notification.Type, notification interface{}) error {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis.demo.svc:6379", // Redis server address
		Password: "",                    // No password
		DB:       0,                     // Default DB
	})
	defer client.Close()

	// Subscribe to a Redis channel
	pubsub := client.Subscribe(context.TODO(), "notifications")
	defer pubsub.Close()

	return client.Publish(context.TODO(), "notifications", notification).Err()
}
