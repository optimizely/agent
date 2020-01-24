package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/optimizely/agent/pkg/middleware"
)

type OverrideBody struct {
	userId       string
	variationKey string
}

// UserOverrideBody - set a forced variation
func Override(w http.ResponseWriter, r *http.Request) {
	optlyClient, optlyContext, err := parseContext(r)
	if err != nil {
		RenderError(err, http.StatusInternalServerError, w, r)
		return
	}
	experimentKey := chi.URLParam(r, "experimentKey")
	if experimentKey == "" {
		RenderError(errors.New("empty experimentKey"), http.StatusBadRequest, w, r)
		return
	}

	override := &UserOverrideBody{}
	if err = ParseRequestBody(r, override); err != nil {
		RenderError(errors.New("empty variationKey"), http.StatusBadRequest, w, r)
		return
	}

	wasSet, err := optlyClient.SetForcedVariation(experimentKey, optlyContext.UserContext.ID, override.VariationKey)
	switch {
	case err != nil:
		middleware.GetLogger(r).Error().Err(err).Msg("error setting forced variation")
		RenderError(err, http.StatusInternalServerError, w, r)

	case wasSet:
		w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
