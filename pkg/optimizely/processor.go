package optimizely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/rs/zerolog/log"
)

const jsonContentType = "application/json"

// SidedoorEventProcessor - forwards events to sidedoor API
type SidedoorEventProcessor struct {
	URL string
}

// ProcessEvent - send event to sidedoor API
func (s *SidedoorEventProcessor) ProcessEvent(event event.UserEvent) {
	jsonValue, _ := json.Marshal(event)
	resp, err := http.Post(s.URL, jsonContentType, bytes.NewBuffer(jsonValue))
	// also check response codes
	// resp.StatusCode == 400 is an error
	success := true

	if err != nil {
		// dispatcherLogger.Error("http.Post failed:", err)

		log.Error().Err(err).Msg("Error sending request")
		success = false
	} else {
		if resp.StatusCode == 204 {
			success = true
		} else {
			fmt.Printf("http.Post invalid response %d", resp.StatusCode)
			success = false
		}
	}

	log.Info().Bool("success", success).Send()
}
