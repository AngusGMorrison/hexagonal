package main

import (
	"fmt"
	"log"
	"os"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres/transferrepo"
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
	pg, err := postgres.New(envConfig.DB)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	defer func() {
		if err := pg.Close(); err != nil {
			logger.Printf("Failed to close database: %v", err)
		}
	}()

	transferRepo, err := transferrepo.New(pg, envConfig.App)
	if err != nil {
		return fmt.Errorf("create transfer Repository: %w", err)
	}

	transferService := controller.NewTransferController(transferRepo)

	// Inject the dependencies into the server.
	server := rest.NewServer(logger, envConfig, transferService)

	return server.Run()
}
