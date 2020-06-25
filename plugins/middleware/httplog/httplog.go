package httplog

import (
	"net/http"

	"github.com/go-chi/httplog"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/plugins/middleware"
)

type HTTPLog struct{}

func (h *HTTPLog) Handler() func(http.Handler) http.Handler {
	return httplog.Handler(log.Logger)
}

func init() {
	middleware.Add("httplog", func() middleware.Middleware {
		return &HTTPLog{}
	})
}
