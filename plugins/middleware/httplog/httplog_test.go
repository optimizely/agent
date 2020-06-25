package httplog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/agent/plugins/middleware"
)

func TestInit(t *testing.T) {
	name := "httplog"
	if mw, ok := middleware.Middlewares[name]; !ok {
		assert.Failf(t, "Plugin not registered", "%s DNE in registry", name)
	} else {
		expected := &HTTPLog{}
		assert.Equal(t, expected, mw())
	}
}

func TestHandler(t *testing.T) {
	httpLog := &HTTPLog{}
	handler := httpLog.Handler()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler(http.NotFoundHandler()).ServeHTTP(w, r)
}
