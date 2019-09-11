package optimizely

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/rs/zerolog/log"
)

const jsonContentType = "application/json"

// SidedoorEventProcessor - sends events to sidedoor API
type SidedoorEventProcessor struct {
	URL string
}

// ProcessEvent - send event to sidedoor API
func (s *SidedoorEventProcessor) ProcessEvent(event event.UserEvent) error {
	jsonValue, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling event")
		return err
	}

	resp, err := http.Post(s.URL, jsonContentType, bytes.NewBuffer(jsonValue))
	resp.Body.Close()

	if err != nil {
		log.Error().Err(err).Msg("Error sending request")
		return err
	}

	return nil
}
