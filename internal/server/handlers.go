package server

import (
	"encoding/json"
	"net/http"
)

func (s *ProxyServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := struct {
		SourceURL string `json:"source_url"`
		TargetURL string `json:"target_url"`
		IsRunning bool   `json:"is_running"`
	}{
		SourceURL: s.sourceURL,
		TargetURL: s.targetURL,
		IsRunning: s.isRunning,
	}

	json.NewEncoder(w).Encode(status)
}
