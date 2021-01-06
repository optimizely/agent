package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/agent/pkg/middleware"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/optimizely/optimizelytest"
)

type GetDatafileTestSuite struct {
	suite.Suite
	oc  *optimizely.OptlyClient
	tc  *optimizelytest.TestClient
	mux *chi.Mux
}

func (suite *GetDatafileTestSuite) ClientCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.OptlyClientKey, suite.oc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Setup Mux
func (suite *GetDatafileTestSuite) SetupTest() {
	testClient := optimizelytest.NewClient()
	testClient.ProjectConfig.Datafile = `{"version":4}`
	optlyClient := &optimizely.OptlyClient{
		OptimizelyClient: testClient.OptimizelyClient,
		ConfigManager:    nil,
		ForcedVariations: testClient.ForcedVariations,
	}

	mux := chi.NewMux()
	mux.With(suite.ClientCtx).Get("/datafile", GetDatafile)

	suite.oc = optlyClient
	suite.tc = testClient
	suite.mux = mux
}

func (suite *GetDatafileTestSuite) TestDatafile() {
	req := httptest.NewRequest("GET", "/datafile", nil)
	rec := httptest.NewRecorder()
	suite.mux.ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	// Unmarshal response
	var actual map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &actual)
	suite.NoError(err)

	var expected map[string]interface{}
	datafile := suite.oc.GetOptimizelyConfig().GetDatafile()
	if err = json.Unmarshal([]byte(datafile), &expected); err != nil {
		suite.NoError(err)
	}

	suite.Equal(expected, actual)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetDatafileTestSuite(t *testing.T) {
	suite.Run(t, new(GetDatafileTestSuite))
}

func TestGetDatafileMissingOptlyCtx(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(GetDatafile).ServeHTTP(rec, req)
	assertError(t, rec, "optlyClient not available", http.StatusInternalServerError)
}
