//go:build integration

// Package integration_test contains integration tests for the hexagonal
// application.
package integration_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/handler/rest"
	server "github.com/angusgmorrison/hexagonal/internal/handler/rest"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/classrepo"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/database"
	"github.com/go-playground/validator/v10"
)

const (
	_serverPort = 7532
	_dbPort     = 5432
	_dbName     = "hexagonal_test"
)

func TestMain(m *testing.M) {
	logger := log.New(os.Stdout, "hexagonal_test ", log.LstdFlags)

	server, err := NewServer(logger)
	if err != nil {
		logger.Fatalf("Create server: %v\n", err)
	}

	go func() {
		if err := server.Run(); err != nil {
			logger.Fatalf("Run server: %v\n", err)
		}
	}()

	defer func() {
		if err := server.GracefulShutdown(); err != nil {
			logger.Fatalf("Shut down server: %v\n", err)
		}
	}()

	os.Exit(m.Run())
}

func NewServer(logger *log.Logger) (*server.Server, error) {
	envConfig := defaultEnvConfig()

	db, err := database.New(envConfig.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	var (
		atomicRepo = classrepo.NewAtomic(db)
		validate   = validator.New()
		service    = classservice.New(logger, validate, atomicRepo)
		server     = rest.NewServer(logger, envConfig, service)
	)

	return server, nil
}

func defaultEnvConfig() envconfig.EnvConfig {
	return envconfig.EnvConfig{
		App: envconfig.App{
			Name: "hexagonal",
			Env:  "test",
			Root: defaultAppRoot(),
		},
		HTTP: envconfig.HTTP{
			Host:                "",
			Port:                _serverPort,
			ClientTimeout:       5 * time.Second,
			ReadTimeout:         5 * time.Second,
			WriteTimeout:        5 * time.Second,
			ShutdownGracePeriod: 0,
		},
		DB: envconfig.DB{
			Host:            "postgres",
			Port:            _dbPort,
			Username:        "postgres",
			Password:        "postgres",
			Name:            _dbName,
			SSLMode:         "disable",
			ConnTimeout:     5 * time.Second,
			ConnMaxIdleTime: 0,
			ConnMaxLifetime: 0,
			MaxIdleConns:    20,
			MaxOpenConns:    20,
		},
	}
}

func defaultAppRoot() string {
	return filepath.Join(string(filepath.Separator), "usr", "src", "app")
}

func enrollmentURL() string {
	return serverURL() + "/enroll"
}

func serverURL() string {
	return fmt.Sprintf("http://0.0.0.0:%d", _serverPort)
}
