package syncer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/rs/zerolog"
)

const (
	PubSubChan  = "optimizely-notifications"
	PubSubRedis = "redis"
)

type Notification struct {
	Type    notification.Type
	Message interface{}
}

type RedisPubSubSyncer struct {
	Host     string
	Password string
	Database int
	Channel  string
	logger   *zerolog.Logger
}

func NewRedisPubSubSyncer(logger *zerolog.Logger, conf *config.SyncConfig) (*RedisPubSubSyncer, error) {
	if !conf.Notification.Enable {
		return nil, errors.New("notification syncer is not enabled")
	}
	if conf.Notification.Default != PubSubRedis {
		return nil, errors.New("redis syncer is not set as default")
	}

	redisConfig, found := conf.Notification.Pubsub[PubSubRedis].(map[string]interface{})
	if !found {
		return nil, errors.New("redis pubsub config not found")
	}

	host, ok := redisConfig["host"].(string)
	if !ok {
		return nil, errors.New("redis host not provided in correct format")
	}
	password, ok := redisConfig["password"].(string)
	if !ok {
		return nil, errors.New("redis password not provider in correct format")
	}
	database, ok := redisConfig["database"].(int)
	if !ok {
		return nil, errors.New("redis database not provided in correct format")
	}
	channel, ok := redisConfig["channel"].(string)
	if !ok {
		return nil, errors.New("redis channel not provided in correct format")
	}

	return &RedisPubSubSyncer{
		Host:     host,
		Password: password,
		Database: database,
		Channel:  channel,
		logger:   logger,
	}, nil
}

func (r *RedisPubSubSyncer) AddHandler(_ notification.Type, _ func(interface{})) (int, error) {
	return 0, nil
}

func (r *RedisPubSubSyncer) RemoveHandler(_ int, t notification.Type) error {
	return nil
}

func (r *RedisPubSubSyncer) Send(t notification.Type, n interface{}) error {
	notification := Notification{
		Type:    t,
		Message: n,
	}
	jsonEvent, err := json.Marshal(notification)
	if err != nil {
		r.logger.Error().Msg("encoding notification to json")
		return err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     r.Host,     // Redis server address
		Password: r.Password, // No password
		DB:       r.Database, // Default DB
	})
	defer client.Close()

	// Subscribe to a Redis channel
	pubsub := client.Subscribe(context.TODO(), r.Channel)
	defer pubsub.Close()

	if err := client.Publish(context.TODO(), r.Channel, jsonEvent).Err(); err != nil {
		r.logger.Err(err).Msg("failed to publish json event to pub/sub")
		return err
	}
	return nil
}
