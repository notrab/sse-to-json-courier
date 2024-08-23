package main

import (
	"log"
	"net/http"
	"os"

	"github.com/notrab/sse-to-json-courier/internal/config"
	"github.com/notrab/sse-to-json-courier/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Error loading configuration: %v", err)
	}

	proxyServer := server.NewProxyServer(logger)
	err = proxyServer.Start(cfg.SourceURL, cfg.TargetURL, cfg.AuthToken)
	if err != nil {
		logger.Fatalf("Error starting proxy: %v", err)
	}

	http.Handle("/", proxyServer)
	logger.Printf("Starting server on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
