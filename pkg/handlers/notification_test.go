/****************************************************************************
 * Copyright 2020,2023, Optimizely, Inc. and contributors                   *
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

// Package handlers //
package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
	"github.com/optimizely/agent/pkg/syncer"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/stretchr/testify/suite"
)

type NotificationTestSuite struct {
	suite.Suite
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

type NotificationMW struct {
	optlyClient *optimizely.OptlyClient
}

func (o *NotificationMW) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, o.optlyClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *NotificationTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	EventStreamMW := &NotificationMW{optlyClient}

	mux.Use(EventStreamMW.ClientCtx)

	suite.mux = mux
	suite.tc = testClient
}

func (suite *NotificationTestSuite) TestFeatureTestFilter() {
	conf := config.NewDefaultConfig()
	suite.mux.Get("/notifications/event-stream", NotificationEventStreamHandler(getMockNotificationReceiver(conf.Synchronization)))

	feature := entities.Feature{Key: "one"}
	suite.tc.AddFeatureTest(feature)

	req := httptest.NewRequest("GET", "/notifications/event-stream?filter="+string(notification.Track)+","+string(notification.ProjectConfigUpdate), nil)
	rec := httptest.NewRecorder()

	expected := ""

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 1*time.Second)

	go func() {
		suite.tc.OptimizelyClient.IsFeatureEnabled(
			"one",
			entities.UserContext{
				ID:                "testUser",
				Attributes:        make(map[string]interface{}),
				QualifiedSegments: make([]string, 0)},
		)
	}()

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
}

func (suite *NotificationTestSuite) TestFilter() {
	filter := []string{"decision", "track"}

	notifications := getFilter(filter)

	suite.True(len(notifications) == 2)
	suite.EqualValues(notification.Track, notifications["track"])
	suite.EqualValues(notification.Decision, notifications["decision"])

	filter = []string{"decision,track", "track"}

	notifications = getFilter(filter)

	suite.True(len(notifications) == 2)
	suite.EqualValues(notification.Track, notifications["track"])
	suite.EqualValues(notification.Decision, notifications["decision"])
}

func (suite *NotificationTestSuite) TestTrackAndProjectConfig() {
	event := entities.Event{Key: "one"}
	suite.tc.AddEvent(event)

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	expected := `data: {"test":"value"}` + "\n\n" + `data: {"Type":"project_config_update","Revision":"revision"}` + "\n\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 3*time.Second)

	nc := registry.GetNotificationCenter("")

	notifications := make([]syncer.Notification, 0)

	trackEvent := map[string]string{"test": "value"}
	projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
		Type:     notification.ProjectConfigUpdate,
		Revision: suite.tc.ProjectConfig.GetRevision(),
	}

	notifications = append(notifications, syncer.Notification{Type: notification.Track, Message: trackEvent})
	notifications = append(notifications, syncer.Notification{Type: notification.ProjectConfigUpdate, Message: projectConfigUpdateNotification})

	go func() {
		time.Sleep(1 * time.Second)

		_ = nc.Send(notification.Track, trackEvent)

		_ = nc.Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification)
	}()

	conf := config.NewDefaultConfig()
	suite.mux.Get("/notifications/event-stream", NotificationEventStreamHandler(getMockNotificationReceiver(conf.Synchronization, notifications...)))

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
}

func (suite *NotificationTestSuite) TestTrackAndProjectConfigWithSynchronization() {
	event := entities.Event{Key: "one"}
	suite.tc.AddEvent(event)

	req := httptest.NewRequest("GET", "/notifications/event-stream", nil)
	rec := httptest.NewRecorder()

	expected := `data: {"test":"value"}` + "\n\n" + `data: {"Type":"project_config_update","Revision":"revision"}` + "\n\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 3*time.Second)

	nc := registry.GetNotificationCenter("")

	notifications := make([]syncer.Notification, 0)

	trackEvent := map[string]string{"test": "value"}
	projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
		Type:     notification.ProjectConfigUpdate,
		Revision: suite.tc.ProjectConfig.GetRevision(),
	}

	notifications = append(notifications, syncer.Notification{Type: notification.Track, Message: trackEvent})
	notifications = append(notifications, syncer.Notification{Type: notification.ProjectConfigUpdate, Message: projectConfigUpdateNotification})

	go func() {
		time.Sleep(1 * time.Second)

		_ = nc.Send(notification.Track, trackEvent)

		_ = nc.Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification)
	}()

	conf := config.NewDefaultConfig()
	conf.Synchronization = config.SyncConfig{
		Notification: config.NotificationConfig{
			Enable:  true,
			Default: "redis",
			Pubsub: map[string]interface{}{
				"redis": map[string]interface{}{
					"host":     "localhost:6379",
					"password": "",
					"database": 0,
				},
			},
		},
	}
	suite.mux.Get("/notifications/event-stream", NotificationEventStreamHandler(getMockNotificationReceiver(conf.Synchronization, notifications...)))

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
}

func (suite *NotificationTestSuite) TestActivateExperimentRaw() {
	testVariation := suite.tc.ProjectConfig.CreateVariation("variation_a")
	suite.tc.AddExperiment("one", []entities.Variation{testVariation})

	req := httptest.NewRequest("GET", "/notifications/event-stream?raw=yes", nil)
	rec := httptest.NewRecorder()

	expected := `{"key":"value"}` + "\n"

	// create a cancelable request context
	ctx := req.Context()
	ctx1, _ := context.WithTimeout(ctx, 2*time.Second)

	nc := registry.GetNotificationCenter("")
	decisionEvent := map[string]string{"key": "value"}

	notifications := make([]syncer.Notification, 0)
	notifications = append(notifications, syncer.Notification{Type: notification.Decision, Message: decisionEvent})

	go func() {
		time.Sleep(1 * time.Second)
		nc.Send(notification.Decision, decisionEvent)
	}()

	conf := config.NewDefaultConfig()
	suite.mux.Get("/notifications/event-stream", NotificationEventStreamHandler(getMockNotificationReceiver(conf.Synchronization, notifications...)))

	suite.mux.ServeHTTP(rec, req.WithContext(ctx1))

	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	response := string(rec.Body.Bytes())
	suite.Equal(expected, response)
}

func (suite *NotificationTestSuite) assertError(rec *httptest.ResponseRecorder, msg string, code int) {
	assertError(suite.T(), rec, msg, code)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEventStreamTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}

func TestEventStreamMissingOptlyCtx(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req := httptest.NewRequest("GET", "/", nil)
	mw := new(NotificationMW)
	mw.optlyClient = nil

	conf := config.NewDefaultConfig()
	handlers := []func(w http.ResponseWriter, r *http.Request){
		NotificationEventStreamHandler(getMockNotificationReceiver(conf.Synchronization)),
	}

	for _, handler := range handlers {
		rec := httptest.NewRecorder()
		mw.ClientCtx(http.HandlerFunc(handler)).ServeHTTP(rec, req)
		assertError(t, rec, "optlyClient not available", http.StatusUnprocessableEntity)
	}
}

func getMockNotificationReceiver(conf config.SyncConfig, msg ...syncer.Notification) NotificationReceiverFunc {
	return func(ctx context.Context) (<-chan syncer.Notification, error) {
		dataChan := make(chan syncer.Notification)
		go func() {
			time.Sleep(1)
			for _, val := range msg {
				dataChan <- val
			}
		}()
		return dataChan, nil
	}
}

func TestDefaultNotificationReceiver(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    <-chan syncer.Notification
		wantErr bool
	}{
		{
			name:    "Test happy path",
			args:    args{ctx: context.WithValue(context.TODO(), SDKKey, "1221")},
			want:    make(chan syncer.Notification),
			wantErr: false,
		},
		{
			name:    "Test without sdk key",
			args:    args{ctx: context.TODO()},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultNotificationReceiver(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultNotificationReceiver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(tt.want) != reflect.TypeOf(got) {
				t.Errorf("DefaultNotificationReceiver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisNotificationReceiver(t *testing.T) {
	conf := config.SyncConfig{
		Notification: config.NotificationConfig{
			Enable:  true,
			Default: "redis",
			Pubsub: map[string]interface{}{
				"redis": map[string]interface{}{
					"host":     "localhost:6379",
					"password": "",
					"database": 0,
				},
			},
		},
	}
	type args struct {
		conf config.SyncConfig
	}
	tests := []struct {
		name string
		args args
		want NotificationReceiverFunc
	}{
		{
			name: "Test happy path",
			args: args{conf: conf},
			want: func(ctx context.Context) (<-chan syncer.Notification, error) {
				return make(<-chan syncer.Notification), nil
			},
		},
		{
			name: "Test empty config",
			args: args{conf: config.SyncConfig{}},
			want: func(ctx context.Context) (<-chan syncer.Notification, error) {
				return nil, errors.New("error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedisNotificationReceiver(tt.args.conf)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("RedisNotificationReceiver() = %v, want %v", got, tt.want)
			}

			ch1, err1 := got(context.TODO())
			ch2, err2 := tt.want(context.TODO())

			if reflect.TypeOf(err1) != reflect.TypeOf(err2) {
				t.Errorf("error type not matched")
			}

			if reflect.TypeOf(ch1) != reflect.TypeOf(ch2) {
				t.Errorf("error type not matched")
			}
		})
	}
}
