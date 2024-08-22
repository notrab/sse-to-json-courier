package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/notrab/sse-to-json-courier/cmd/cli/flags"
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

	port := fmt.Sprintf(":%s", flags.Port)
	go func() {
		logger.Printf("Starting server on %s", port)
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	fmt.Printf("Server started successfully on port %s. Press Ctrl+C to stop.", port)

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
