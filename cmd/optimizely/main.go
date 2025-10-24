/****************************************************************************
 * Copyright 2019,2022-2024 Optimizely, Inc. and contributors               *
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
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v2"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/handlers"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/routers"
	"github.com/optimizely/agent/pkg/server"
	_ "github.com/optimizely/agent/plugins/cmabcache/all"          // Initiate the loading of the cmabCache plugins
	_ "github.com/optimizely/agent/plugins/interceptors/all"       // Initiate the loading of the userprofileservice plugins
	_ "github.com/optimizely/agent/plugins/odpcache/all"           // Initiate the loading of the odpCache plugins
	_ "github.com/optimizely/agent/plugins/userprofileservice/all" // Initiate the loading of the interceptor plugins
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// Version holds the admin version
var Version string // default set at compile time

func initConfig(v *viper.Viper) error {
	// Set explicit defaults
	v.SetDefault("config.filename", "config.yaml") // Configuration file name

	// Load defaults from the AgentConfig by loading the marshaled values as yaml
	// https://github.com/spf13/viper/issues/188
	defaultConf := config.NewDefaultConfig()
	defaultConf.Version = Version
	b, err := yaml.Marshal(defaultConf)
	if err != nil {
		return err
	}

	dc := bytes.NewReader(b)
	v.SetConfigType("yaml")
	return v.MergeConfig(dc)
}

func loadConfig(v *viper.Viper) *config.AgentConfig {
	// Configure environment variables
	v.SetEnvPrefix("optimizely")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read configuration from file
	configFile := v.GetString("config.filename")
	v.SetConfigFile(configFile)
	if err := v.MergeInConfig(); err != nil {
		log.Info().Err(err).Msg("Skip loading configuration from config file.")
	}

	conf := &config.AgentConfig{}
	if err := v.Unmarshal(conf); err != nil {
		log.Info().Err(err).Msg("Unable to marshal configuration.")
	}

	// https://github.com/spf13/viper/issues/406
	if interceptors, ok := v.Get("server.interceptors").(map[string]interface{}); ok {
		conf.Server.Interceptors = interceptors
	}

	// Check if JSON string was set using OPTIMIZELY_CLIENT_USERPROFILESERVICE environment variable
	if userProfileService := v.GetStringMap("client.userprofileservice"); len(userProfileService) > 0 {
		conf.Client.UserProfileService = userProfileService
	}

	// Check if JSON string was set using OPTIMIZELY_CLIENT_ODP_SEGMENTSCACHE environment variable
	if odpSegmentsCache := v.GetStringMap("client.odp.segmentsCache"); len(odpSegmentsCache) > 0 {
		conf.Client.ODP.SegmentsCache = odpSegmentsCache
	}

	// Handle CMAB configuration using the same approach as UserProfileService
	// Check for complete CMAB configuration first
	if cmab := v.GetStringMap("cmab"); len(cmab) > 0 {
		if timeout, ok := cmab["requestTimeout"].(string); ok {
			if duration, err := time.ParseDuration(timeout); err == nil {
				conf.CMAB.RequestTimeout = duration
			}
		}
		if cache, ok := cmab["cache"].(map[string]interface{}); ok {
			conf.CMAB.Cache = cache
		}
		if retryConfig, ok := cmab["retryConfig"].(map[string]interface{}); ok {
			if maxRetries, ok := retryConfig["maxRetries"].(float64); ok {
				conf.CMAB.RetryConfig.MaxRetries = int(maxRetries)
			}
			if initialBackoff, ok := retryConfig["initialBackoff"].(string); ok {
				if duration, err := time.ParseDuration(initialBackoff); err == nil {
					conf.CMAB.RetryConfig.InitialBackoff = duration
				}
			}
			if maxBackoff, ok := retryConfig["maxBackoff"].(string); ok {
				if duration, err := time.ParseDuration(maxBackoff); err == nil {
					conf.CMAB.RetryConfig.MaxBackoff = duration
				}
			}
			if backoffMultiplier, ok := retryConfig["backoffMultiplier"].(float64); ok {
				conf.CMAB.RetryConfig.BackoffMultiplier = backoffMultiplier
			}
		}
	}

	// Check for individual map sections
	if cmabCache := v.GetStringMap("cmab.cache"); len(cmabCache) > 0 {
		conf.CMAB.Cache = cmabCache
	}

	if cmabRetryConfig := v.GetStringMap("cmab.retryConfig"); len(cmabRetryConfig) > 0 {
		if maxRetries, ok := cmabRetryConfig["maxRetries"].(int); ok {
			conf.CMAB.RetryConfig.MaxRetries = maxRetries
		} else if maxRetries, ok := cmabRetryConfig["maxRetries"].(float64); ok {
			conf.CMAB.RetryConfig.MaxRetries = int(maxRetries)
		}
		if initialBackoff, ok := cmabRetryConfig["initialBackoff"].(string); ok {
			if duration, err := time.ParseDuration(initialBackoff); err == nil {
				conf.CMAB.RetryConfig.InitialBackoff = duration
			}
		}
		if maxBackoff, ok := cmabRetryConfig["maxBackoff"].(string); ok {
			if duration, err := time.ParseDuration(maxBackoff); err == nil {
				conf.CMAB.RetryConfig.MaxBackoff = duration
			}
		}
		if backoffMultiplier, ok := cmabRetryConfig["backoffMultiplier"].(float64); ok {
			conf.CMAB.RetryConfig.BackoffMultiplier = backoffMultiplier
		}
	}

	return conf
}

func initLogging(conf config.LogConfig) {
	if conf.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set whether or not the SDK key is included in the logging output of agent and go-sdk
	optimizely.ShouldIncludeSDKKey = conf.IncludeSDKKey
	logging.IncludeSDKKeyInLogFields(conf.IncludeSDKKey)

	if lvl, err := zerolog.ParseLevel(conf.Level); err != nil {
		log.Warn().Err(err).Msg("Error parsing log level")
	} else {
		log.Logger = log.Logger.Level(lvl)
	}
}

func getStdOutTraceProvider(conf config.OTELTracingConfig) (*sdktrace.TracerProvider, error) {
	f, err := os.Create(conf.Services.StdOut.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create the trace file, error: %s", err.Error())
	}

	exp, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(f),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create the collector exporter, error: %s", err.Error())
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(conf.ServiceName),
			semconv.DeploymentEnvironmentKey.String(conf.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create the otel resource, error: %s", err.Error())
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	), nil
}

func getOTELTraceClient(conf config.OTELTracingConfig) (otlptrace.Client, error) {
	switch conf.Services.Remote.Protocol {
	case config.TracingRemoteProtocolHTTP:
		return otlptracehttp.NewClient(
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint(conf.Services.Remote.Endpoint),
		), nil
	case config.TracingRemoteProtocolGRPC:
		return otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(conf.Services.Remote.Endpoint),
		), nil
	default:
		return nil, errors.New("unknown remote tracing protocal")
	}
}

func getRemoteTraceProvider(conf config.OTELTracingConfig) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(conf.ServiceName),
			semconv.DeploymentEnvironmentKey.String(conf.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create the otel resource, error: %s", err.Error())
	}

	traceClient, err := getOTELTraceClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create the remote trace client, error: %s", err.Error())
	}

	traceExporter, err := otlptrace.New(context.Background(), traceClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create the remote trace exporter, error: %s", err.Error())
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(conf.Services.Remote.SampleRate))),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	), nil
}

func initTracing(conf config.OTELTracingConfig) (*sdktrace.TracerProvider, error) {
	switch conf.Default {
	case config.TracingServiceTypeRemote:
		return getRemoteTraceProvider(conf)
	case config.TracingServiceTypeStdOut:
		return getStdOutTraceProvider(conf)
	default:
		return nil, errors.New("unknown tracing service type")
	}
}

func setRuntimeEnvironment(conf config.RuntimeConfig) {
	if conf.BlockProfileRate != 0 {
		log.Warn().Msgf("Setting non-zero blockProfileRate is NOT recommended for production")
		runtime.SetBlockProfileRate(conf.BlockProfileRate)
	}

	if conf.MutexProfileFraction != 0 {
		log.Warn().Msgf("Setting non-zero mutexProfileFraction is NOT recommended for production")
		runtime.SetMutexProfileFraction(conf.MutexProfileFraction)
	}
}

func main() {
	v := viper.New()
	if err := initConfig(v); err != nil {
		log.Panic().Err(err).Msg("Unable to initialize config")
	}

	conf := loadConfig(v)
	initLogging(conf.Log)

	if conf.Tracing.Enabled {
		tp, err := initTracing(conf.Tracing.OpenTelemetry)
		if err != nil {
			log.Panic().Err(err).Msg("Unable to initialize tracing")
		}
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Error().Err(err).Msg("Failed to shutdown tracing")
			}
		}()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		log.Info().Msg(fmt.Sprintf("Tracing enabled with service %q", conf.Tracing.OpenTelemetry.Default))
	} else {
		log.Info().Msg("Tracing disabled")
	}

	conf.LogConfigWarnings()

	setRuntimeEnvironment(conf.Runtime)

	// Set metrics type to be used
	agentMetricsRegistry := metrics.NewRegistry(conf.Admin.MetricsType)
	sdkMetricsRegistry := optimizely.NewRegistry(agentMetricsRegistry)

	ctx, cancel := context.WithCancel(context.Background()) // Create default service context
	defer cancel()
	ctx = context.WithValue(ctx, handlers.LoggerKey, &log.Logger)

	sg := server.NewGroup(ctx, conf.Server) // Create a new server group to manage the individual http listeners
	var tracer trace.Tracer
	if conf.Tracing.Enabled {
		tracer = otel.GetTracerProvider().Tracer(conf.Tracing.OpenTelemetry.ServiceName)
	}
	optlyCache := optimizely.NewCache(ctx, *conf, sdkMetricsRegistry, tracer)
	optlyCache.Init(conf.SDKKeys)

	// goroutine to check for signals to gracefully shutdown listeners
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		// Wait for signal
		sig := <-signalChannel
		log.Info().Msgf("Received signal: %s\n", sig)
		cancel()
	}()

	apiRouter := routers.NewDefaultAPIRouter(optlyCache, *conf, agentMetricsRegistry)
	adminRouter := routers.NewAdminRouter(*conf)

	log.Info().Str("version", conf.Version).Msg("Starting services.")
	sg.GoListenAndServe("api", conf.API.Port, apiRouter)
	sg.GoListenAndServe("webhook", conf.Webhook.Port, routers.NewWebhookRouter(ctx, optlyCache, *conf))
	sg.GoListenAndServe("admin", conf.Admin.Port, adminRouter) // Admin should be added last.

	// wait for server group to shutdown
	if err := sg.Wait(); err == nil {
		log.Info().Msg("Exiting.")
	} else {
		log.Fatal().Err(err).Msg("Exiting.")
	}

	optlyCache.Wait()
}
