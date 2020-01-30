package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

func DecisionCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		optlyClient, err := GetOptlyClient(r)
		logger := GetLogger(r)
		if err != nil {
			RenderError(fmt.Errorf("optlyClient not available in DecisionCtx"), http.StatusInternalServerError, w, r)
			return
		}

		decisionKey := chi.URLParam(r, "decisionKey")
		if decisionKey == "" {
			log.Debug().Msg("no decisionKey provided")
			RenderError(fmt.Errorf("invalid request, missing decisionKey in DecisionCtx"), http.StatusBadRequest, w, r)
			return
		}

		oConf := optlyClient.GetOptimizelyConfig()

		if f, ok := oConf.FeaturesMap[decisionKey]; ok {
			logger.Debug().Str("featureKey", decisionKey).Msg("Added feature to request context in DecisionCtx")
			ctx := context.WithValue(r.Context(), OptlyFeatureKey, &f)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if e, ok := oConf.ExperimentsMap[decisionKey]; ok {
			logger.Debug().Str("experimentKey", decisionKey).Msg("Added experiment to request context in DecisionCtx")
			ctx := context.WithValue(r.Context(), OptlyExperimentKey, &e)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		RenderError(fmt.Errorf("unable to find entity for key %s", decisionKey), http.StatusNotFound, w, r)
		return
	})
}
