package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

	scanner := bufio.NewScanner(resp.Body)
	var event Event
	var eventBuffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || line == ":keep-alive" {
			continue
		}

		eventBuffer.WriteString(line + "\n")

		if strings.HasPrefix(line, "data:") {
			event = parseSSEEvent(eventBuffer.String())
			s.logger.Printf("Received complete event: %+v", event)

			if event.Event == "changes" {
				s.logger.Println("Received 'changes' event, attempting to forward")
				err := s.forwardEvent(event.Data)
				if err != nil {
					s.logger.Printf("Error forwarding event: %v", err)
				}
			} else if event.Event == "error" {
				s.logger.Printf("Received error event: %s", string(event.Data))
			}

			eventBuffer.Reset()
			event = Event{}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SSE stream: %v", err)
	}

	return nil
}

func (s *ProxyServer) forwardEvent(eventData json.RawMessage) error {
	var changeEvent struct {
		Insert int `json:"insert"`
		Update int `json:"update"`
		Delete int `json:"delete"`
	}
	err := json.Unmarshal(eventData, &changeEvent)
	if err != nil {
		return fmt.Errorf("error parsing change event: %v", err)
	}

	s.logger.Printf("Attempting to forward event: %+v", changeEvent)

	targetURL, err := url.Parse(s.targetURL)
	if err != nil {
		return fmt.Errorf("error parsing target URL: %v", err)
	}

	req, err := http.NewRequest("POST", targetURL.String(), bytes.NewReader(eventData))
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
