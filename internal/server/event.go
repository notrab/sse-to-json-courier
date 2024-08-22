package server

import (
	"encoding/json"
	"strings"
)

type Event struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ChangeEvent struct {
	Table   string          `json:"table"`
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func parseSSEEvent(eventString string) Event {
	var event Event
	lines := strings.Split(strings.TrimSpace(eventString), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			event.Data = json.RawMessage(dataStr)
		}
	}

	return event
}
