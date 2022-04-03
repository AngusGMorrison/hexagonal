// Package integration contains integration tests for the hexagonal application.
package integration

import (
	"testing"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
)

func TestMain(m *testing.M) {
	env := defaultEnvConfig()
	
}

func defaultEnvConfig() envconfig.EnvConfig {
	return envconfig.EnvConfig{
		App: envconfig.App{
			Name: "hexagonal",
			Env:  "test",
			Root: "usr/src/app",
		},
		HTTP: envconfig.HTTP{
			Host:                "",
			Port:                3000,
			ReadTimeout:         5 * time.Second,
			WriteTimeout:        5 * time.Second,
			ShutdownGracePeriod: 0,
		},
		DB: envconfig.DB{
			Host:            "postgres",
			Port:            5432,
			Username:        "postgres",
			Password:        "postgres",
			Name:            "hexagonal_test",
			SSLMode:         "disable",
			ConnTimeout:     0,
			ConnMaxIdleTime: 0,
			ConnMaxLifetime: 0,
			MaxIdleConns:    20,
			MaxOpenConns:    20,
		},
	}
}
