package main

import (
	"log"
	"net/http"
	"os"

	"github.com/notrab/sse-to-json-courier/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	proxyServer := server.NewProxyServer(logger)

	http.Handle("/", proxyServer)
	logger.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
