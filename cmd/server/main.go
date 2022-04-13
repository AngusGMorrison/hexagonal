package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
	"github.com/angusgmorrison/hexagonal/internal/adapter/rest"
	"github.com/angusgmorrison/hexagonal/internal/controller"
)

func main() {
	logger := log.New(os.Stdout, "hexagonal_migrate ", log.LstdFlags)

	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger *log.Logger) error {
	// Load environment variables.
	envConfig, err := envconfig.New()
	if err != nil {
		return fmt.Errorf("create envconfig: %w", err)
	}

	// Set up the server's IO dependencies.
	db, err := postgres.NewDB(envConfig.DB)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Printf("Failed to close database: %v", err)
		}
	}()

	transferQueryDir := filepath.Join(envConfig.App.Root, postgres.RelativeQueryDir(), "transfer")
	repo, err := postgres.NewTransferRepository(db, transferQueryDir)
	if err != nil {
		return fmt.Errorf("create TransferRepository: %w", err)
	}

	transferController := controller.NewTransferController(logger, repo)

	// Inject the dependencies into the server.
	server := rest.NewServer(logger, envConfig, transferController)

	return server.Run()
}
