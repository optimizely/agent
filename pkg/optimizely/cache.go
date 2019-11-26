/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package optimizely wraps the Optimizely SDK
package optimizely

import (
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/sidedoor/pkg/event"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decision"
	events "github.com/optimizely/go-sdk/pkg/event"
	cmap "github.com/orcaman/concurrent-map"
	snsq "github.com/segmentio/nsq-go"
)

// EPQSize integer event processor queue size
const EPQSize = "optimizely.eventProcessor.queueSize"
// EPBSize integer event processor batch size
const EPBSize = "optimizely.eventProcessor.batchSize"
// NSQEnable boolean true enables using the NSQ as the queue for the event processor
const NSQEnabled = "optimizely.eventProcessor.nsq.enabled"
// NSQStartEmbedded boolean whether to start the embedded nsq daemon
const NSQStartEmbedded = "optimizely.eventProcessor.nsq.startEmbedded"
// NSQAddress string address to bind the consumer and/or producer
const NSQAddress = "optimizely.eventProcessor.nsq.address"
// NSQConsumer boolean.  Start the consumer if set to true
const NSQConsumer = "optimizely.eventProcessor.nsq.withConsumer"
// NSQProducer boolan.  Start the producer if set to true
const NSQProducer = "optimizely.eventProcessor.nsq.withProducer"

// OptlyCache implements the Cache interface backed by a concurrent map.
// The default OptlyClient lookup is based on supplied configuration via env variables.
type OptlyCache struct {
	loader   func(string) (*OptlyClient, error)
	optlyMap cmap.ConcurrentMap
}

// NewCache returns a new implementation of OptlyCache interface backed by a concurrent map.
func NewCache() *OptlyCache {
	cache := &OptlyCache{
		optlyMap: cmap.New(),
		loader:   initOptlyClient,
	}

	cache.init()
	return cache
}

func (c *OptlyCache) init() {
	sdkKeys := viper.GetStringSlice("optimizely.sdkKeys")
	for _, sdkKey := range sdkKeys {
		if _, err := c.GetClient(sdkKey); err != nil {
			log.Warn().Str("sdkKey", sdkKey).Msg("Failed to initialize Opimizely Client.")
		}
	}
}

// GetClient is used to fetch an instance of the OptlyClient when the SDK Key is explicitly supplied.
func (c *OptlyCache) GetClient(sdkKey string) (*OptlyClient, error) {
	val, ok := c.optlyMap.Get(sdkKey)
	if ok {
		return val.(*OptlyClient), nil
	}

	oc, err := c.loader(sdkKey)
	if err != nil {
		return oc, err
	}

	set := c.optlyMap.SetIfAbsent(sdkKey, oc)
	if set {
		return oc, err
	}

	// If we didn't "set" the key in this method execution then it was set in another thread.
	// Recursively lookuping up the SDK key "should" only happen once.
	return c.GetClient(sdkKey)
}

// GetOptlyEventProcessor get the optly event processor using viper configuration variables.
func GetOptlyEventProcessor() events.Processor {
	var ep events.Processor
	batchSize := events.DefaultBatchSize
	queueSize := events.DefaultEventQueueSize

	var q events.Queue

	if viper.IsSet(NSQEnabled) && viper.GetBool(NSQEnabled) {
		startEmbedded := viper.IsSet(NSQStartEmbedded) && viper.GetBool(NSQStartEmbedded)
		var nsqAddress string
		if nsqAddress = viper.GetString(NSQAddress); nsqAddress == "" {
			nsqAddress = event.NsqListenSpec
		}
		withProducer := viper.GetBool(NSQProducer)
		withConsumer := viper.GetBool(NSQConsumer)
		if viper.IsSet(EPQSize) {
			queueSize = viper.GetInt(EPQSize)
		}

		var consumer *snsq.Consumer
		var producer *snsq.Producer

		if withConsumer {
			consumer, _ = snsq.StartConsumer(snsq.ConsumerConfig{
				Topic:       event.NsqTopic,
				Channel:     event.NsqConsumerChannel,
				Address:     nsqAddress,
				MaxInFlight: 100,
			})

		}

		if withProducer {
			nsqConfig := snsq.ProducerConfig{Address: nsqAddress, Topic: event.NsqTopic}
			producer, _ = snsq.NewProducer(nsqConfig)

		}

		q, _ = event.NewNSQueue(queueSize, startEmbedded, producer, consumer)

	} else {
		q = events.NewInMemoryQueue(queueSize)
	}

	if viper.IsSet(EPQSize) || viper.IsSet(EPBSize) {
		if viper.IsSet(EPQSize) {
			queueSize = viper.GetInt(EPQSize)
		}
		if viper.IsSet(EPBSize) {
			batchSize = viper.GetInt(EPBSize)
		}
	}


	ep = events.NewBatchEventProcessor(events.WithQueueSize(queueSize), events.WithBatchSize(batchSize), events.WithQueue(q))

	return ep
}

func initOptlyClient(sdkKey string) (*OptlyClient, error) {
	log.Info().Str("sdkKey", sdkKey).Msg("Loading Optimizely instance")
	configManager := config.NewPollingProjectConfigManager(sdkKey)
	if _, err := configManager.GetConfig(); err != nil {
		return &OptlyClient{}, err
	}

	ep := GetOptlyEventProcessor()

	epFun := client.WithEventProcessor(ep)
	if ep == nil {
		epFun = nil
	}

	forcedVariations := decision.NewMapExperimentOverridesStore()
	optimizelyFactory := &client.OptimizelyFactory{}
	optimizelyClient, err := optimizelyFactory.Client(
		client.WithConfigManager(configManager),
		client.WithExperimentOverrides(forcedVariations),
		epFun,
	)
	return &OptlyClient{optimizelyClient, configManager, forcedVariations}, err
}
