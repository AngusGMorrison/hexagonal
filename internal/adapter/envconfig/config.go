// Package envconfig loads environment variables and exposes them to the
// application.
package envconfig

import (
	"fmt"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// envVarPrefix defines a common prefix for environment variables.
const envVarPrefix = ""

// New loads all environment variables and returns them as an EnvConfig object.
func New() (EnvConfig, error) {
	var env EnvConfig
	if err := envconfig.Process(envVarPrefix, &env); err != nil {
		return EnvConfig{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return env, nil
}

// Env represents the environment variables of the running application.
type EnvConfig struct {
	App  App
	HTTP HTTP
	DB   DB
}

// App represents environment variables related to the identity and general
// function of the application.
type App struct {
	Name string `envconfig:"APP_NAME" default:"hexagonal"`
	Env  string `envconfig:"APP_ENV" required:"true"`
}

// HTTP represents all HTTP-related environment variables.
type HTTP struct {
	Host                string        `envconfig:"SERVER_HOST" default:""`
	Port                int           `envconfig:"SERVER_PORT" default:"3000"`
	ReadTimeout         time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"5s"`
	WriteTimeout        time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"5s"`
	ShutdownGracePeriod time.Duration `envconfig:"SERVER_SHUTDOWN_GRACE_PERIOD" default:"0s"`
}

// DB represents all DB-related environment variables.
type DB struct {
	Host            string        `envconfig:"DB_HOST" required:"true"`
	Port            int           `envconfig:"DB_PORT" required:"true"`
	Username        string        `envconfig:"DB_USERNAME" required:"true"`
	Password        string        `envconfig:"DB_PASSWORD" required:"true"`
	Name            string        `envconfig:"DB_NAME" required:"true"`
	ConnTimeout     time.Duration `envconfig:"DB_CONN_TIMEOUT" default:"5s"`
	ConnMaxIdleTime time.Duration `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"0s"`
	ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"0s"`
	MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"20"`
	MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"20"`
	SSLMode         string        `envconfig:"DB_SSL_MODE" default:"require"`
}

// URL returns the URL of the database.
func (db DB) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=UTC",
		db.Username, url.QueryEscape(db.Password), db.Host, db.Port, db.Name, db.SSLMode)
}
