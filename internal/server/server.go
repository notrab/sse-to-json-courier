package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type ProxyServer struct {
	sourceURL  string
	targetURL  string
	authToken  string
	httpClient *http.Client
	logger     *log.Logger
	mu         sync.Mutex
	isRunning  bool
}

type Event struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ChangeEvent struct {
	Table   string          `json:"table"`
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func NewProxyServer(logger *log.Logger) *ProxyServer {
	return &ProxyServer{
		httpClient: &http.Client{},
		logger:     logger,
	}
}

func (s *ProxyServer) Start(sourceURL, targetURL, authToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	s.sourceURL = sourceURL
	s.targetURL = targetURL
	s.authToken = authToken

	if s.sourceURL == "" || s.targetURL == "" {
		return fmt.Errorf("source URL and target URL must be set")
	}

	go s.connectToSSE()
	s.isRunning = true
	return nil
}

func (s *ProxyServer) connectToSSE() {
	for {
		err := s.streamSSE()
		if err != nil {
			s.logger.Printf("Error in SSE connection: %v. Reconnecting in 5 seconds...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *ProxyServer) streamSSE() error {
	req, err := http.NewRequest("GET", s.sourceURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	if s.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.authToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error connecting to SSE source: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	s.logger.Printf("Connected to SSE source successfully")

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("SSE stream closed")
			}
			return fmt.Errorf("error reading SSE: %v", err)
		}

		if len(line) == 1 && line[0] == '\n' {
			continue
		}

		if strings.HasPrefix(string(line), ":keep-alive") {
			s.logger.Println("Received keep-alive (skipping)")
			continue
		}

		s.logger.Printf("Received event: %s", string(line))

		event, err := parseSSEEvent(line)
		if err != nil {
			s.logger.Printf("Error parsing SSE event: %v", err)
			continue
		}

		if event.Event == "changes" {
			s.logger.Println("Received 'changes' event, attempting to forward")
			err := s.forwardEvent(event.Data)
			if err != nil {
				s.logger.Printf("Error forwarding event: %v", err)
			}
		} else if event.Event == "error" {
			s.logger.Printf("Received error event: %s", string(event.Data))
		} else {
			s.logger.Printf("Received unknown event type: %s", event.Event)
		}
	}
}

func parseSSEEvent(data []byte) (Event, error) {
	var event Event
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return event, fmt.Errorf("invalid SSE format")
	}

	event.Event = strings.TrimSpace(parts[0])
	event.Data = json.RawMessage(strings.TrimSpace(parts[1]))
	return event, nil
}

func (s *ProxyServer) forwardEvent(eventData json.RawMessage) error {
	var changeEvent ChangeEvent
	err := json.Unmarshal(eventData, &changeEvent)
	if err != nil {
		return fmt.Errorf("error parsing change event: %v", err)
	}

	s.logger.Printf("Attempting to forward event: table=%s, action=%s", changeEvent.Table, changeEvent.Action)

	targetURL, err := url.Parse(s.targetURL)
	if err != nil {
		return fmt.Errorf("error parsing target URL: %v", err)
	}

	query := targetURL.Query()
	query.Set("table", changeEvent.Table)
	query.Set("action", changeEvent.Action)
	targetURL.RawQuery = query.Encode()

	req, err := http.NewRequest("POST", targetURL.String(), bytes.NewReader(changeEvent.Payload))
	if err != nil {
		s.logger.Printf("Failed to create forward request: %v", err)
		return fmt.Errorf("error creating forward request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	s.logger.Printf("Forwarding to: %s", targetURL.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Printf("Failed to forward event: %v", err)

		return fmt.Errorf("error forwarding event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Printf("Successfully forwarded event: POST %s - %d", targetURL.String(), resp.StatusCode)
	} else {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Printf("Failed to forward event: POST %s - %d, response: %s", targetURL.String(), resp.StatusCode, string(body))
		return fmt.Errorf("unexpected status code when forwarding: %d", resp.StatusCode)
	}

	return nil
}

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.HandleFunc("/status", s.handleStatus).Methods("GET")

	router.ServeHTTP(w, r)
}

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
