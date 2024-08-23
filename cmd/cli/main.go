package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/notrab/sse-to-json-courier/cmd/cli/flags"
	"github.com/notrab/sse-to-json-courier/internal/config"
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
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SOURCE_URL=%s", flags.SourceURL),
		fmt.Sprintf("TARGET_URL=%s", flags.TargetURL),
		fmt.Sprintf("AUTH_TOKEN=%s", flags.AuthToken),
		fmt.Sprintf("PORT=%s", flags.Port),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	fmt.Printf("Server started successfully on port %s. Press Ctrl+C to stop.\n", flags.Port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down server...")
	cmd.Process.Kill()
}

func main() {
	if len(os.Args) == 1 {
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
	} else {
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
