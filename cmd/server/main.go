package main

import (
	"fmt"
	"log"
	"os"

	"github.com/angusgmorrison/hexagonal/internal/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/handler/rest"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/classrepo"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/database"
	"github.com/go-playground/validator/v10"
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
	db, err := database.New(envConfig.DB)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Printf("Failed to close database: %v", err)
		}
	}()

	var (
		validate     = validator.New()
		classRepo    = classrepo.NewAtomic(db)
		classService = classservice.New(logger, validate, classRepo)
		server       = rest.NewServer(logger, envConfig, classService)
	)

	return server.Run()
}
