package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/notrab/sse-to-json-courier/internal/flags"
	"github.com/notrab/sse-to-json-courier/internal/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sse-proxy",
	Short: "SSE Proxy is a server that forwards Server-Sent Events",
	Long:  `SSE Proxy is a server that listens to a source URL for Server-Sent Events and forwards them to a target URL.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the SSE Proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	flags.AddFlags(startCmd)
	rootCmd.AddCommand(startCmd)
}

func startServer() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	proxyServer := server.NewProxyServer(logger)

	err := proxyServer.Start(flags.SourceURL, flags.TargetURL, flags.AuthToken)
	if err != nil {
		log.Fatalf("Error starting proxy: %v", err)
	}

	http.Handle("/", proxyServer)
	go func() {
		logger.Println("Starting server on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	fmt.Println("Server started successfully. Press Ctrl+C to stop.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down server...")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
