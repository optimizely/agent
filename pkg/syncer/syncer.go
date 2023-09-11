package syncer

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/rs/zerolog"
)

const (
	PubSubChan  = "optimizely-notifications"
	PubSubRedis = "redis"
)

type RedisCenter struct {
	Addr     string
	Password string
	DB       int
	logger   *zerolog.Logger
}

// AddHandler(Type, func(interface{})) (int, error)
// RemoveHandler(int, Type) error
// Send(Type, interface{}) error

func NewRedisCenter(conf *config.SyncConfig) *RedisCenter {
	return &RedisCenter{
		Addr:     conf.Notification.Pubsub.Addr,
		Password: conf.Notification.Pubsub.Password,
		DB:       conf.Notification.Pubsub.DB,
	}
}

func (r *RedisCenter) AddHandler(_ notification.Type, _ func(interface{})) (int, error) {
	return 0, nil
}

func (r *RedisCenter) RemoveHandler(_ int, t notification.Type) error {
	return nil
}

func (r *RedisCenter) Send(_ notification.Type, n interface{}) error {
	jsonEvent, err := json.Marshal(n)
	if err != nil {
		r.logger.Error().Msg("encoding notification to json")
		return err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     r.Addr,     // Redis server address
		Password: r.Password, // No password
		DB:       r.DB,       // Default DB
	})
	defer client.Close()

	// Subscribe to a Redis channel
	pubsub := client.Subscribe(context.TODO(), PubSubChan)
	defer pubsub.Close()

	if err := client.Publish(context.TODO(), PubSubChan, jsonEvent).Err(); err != nil {
		r.logger.Err(err).Msg("failed to publish json event to pub/sub")
		return err
	}
	return nil
}

type RedisPubSubSyncer struct {
	Addr     string
	Password string
	DB       int
	logger   *zerolog.Logger
}

func NewRedisPubSubSyncer(logger *zerolog.Logger, conf *config.SyncConfig) *RedisPubSubSyncer {
	return &RedisPubSubSyncer{
		Addr:     conf.Notification.Pubsub.Addr,
		Password: conf.Notification.Pubsub.Password,
		DB:       conf.Notification.Pubsub.DB,
		logger:   logger,
	}
}

func (r *RedisPubSubSyncer) GetNotificationSyncer(ctx context.Context) func(n interface{}) {
	return func(n interface{}) {
		jsonEvent, err := json.Marshal(n)
		if err != nil {
			r.logger.Error().Msg("encoding notification to json")
			return
		}
		client := redis.NewClient(&redis.Options{
			Addr:     r.Addr,     // Redis server address
			Password: r.Password, // No password
			DB:       r.DB,       // Default DB
		})
		defer client.Close()

		// Subscribe to a Redis channel
		pubsub := client.Subscribe(ctx, PubSubChan)
		defer pubsub.Close()

		if err := client.Publish(ctx, PubSubChan, jsonEvent).Err(); err != nil {
			r.logger.Err(err).Msg("failed to publish json event to pub/sub")
			return
		}
	}
}
