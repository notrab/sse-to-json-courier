package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

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

// todo: implement NewProxyServer flags
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

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	router.HandleFunc("/status", s.handleStatus).Methods("GET")

	router.ServeHTTP(w, r)
}
