package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/tushariitr-19/patents-mcp/logger"
	"github.com/tushariitr-19/patents-mcp/server"
)

func main() {
	debug := os.Getenv("DEBUG") == "true"
	if err := logger.Init(debug); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	if os.Getenv("GCP_PROJECT_ID") == "" {
		logger.Log.Fatal("GCP_PROJECT_ID environment variable is required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Log.Info("starting patents-mcp server", zap.String("version", "v0.1.0"))

	s := server.New()
	if err := s.Run(ctx); err != nil {
		logger.Log.Error("server stopped", zap.Error(err))
	}
}
