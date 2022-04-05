// Package server specifies an http server capable of receiving bulk transfer
// requests.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/http/server/handler"
	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain"
	"github.com/gin-gonic/gin"
)

// ErrServerClosed wraps http.ErrServerClosed, preventing external packages from
// needing to require package http in addition to package server.
var ErrServerClosed = http.ErrServerClosed

// EnvNotSupportedError represents a failure to match an environment string
// passed in a server's Config.
type EnvNotSupportedError struct {
	env string
}

func (e EnvNotSupportedError) Error() string {
	return fmt.Sprintf("unsupported env %q", e.env)
}

// Server wraps a standard library server, providing support for
// application-specific methods.
type Server struct {
	server *http.Server

	log         *log.Logger
	config      envconfig.HTTP
	errorStream chan error
}

// Config provides a mechanism for required server configuration that avoids us
// having to refactor function signatures when the configuration changes.
type Config struct {
	Env             envconfig.EnvConfig
	Logger          *log.Logger
	TransferService *transferdomain.Service
}

// New returns a new hexagonal server configured using the provided Config.
func New(cfg Config) (*Server, error) {
	handler, err := newHandler(cfg)
	if err != nil {
		return nil, err
	}

	server := Server{
		server: &http.Server{
			Addr:         serverAddress(cfg.Env.HTTP.Host, cfg.Env.HTTP.Port),
			ReadTimeout:  cfg.Env.HTTP.ReadTimeout,
			WriteTimeout: cfg.Env.HTTP.WriteTimeout,
			Handler:      handler,
		},
		log:         cfg.Logger,
		config:      cfg.Env.HTTP,
		errorStream: make(chan error, 1),
	}

	return &server, nil
}

// Run starts the Server in a new goroutine, forwarding any errors to its
// errorStream.
func (s *Server) Run() {
	s.log.Printf("Starting server at %s\n", s.server.Addr)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.errorStream <- fmt.Errorf("ListenAndServe: %w", err)
		}
	}()
}

// Errors returns the Server's errorStream as a receive-only channel that can be
// used to monitor the health of the server across process boundaries.
func (s *Server) Errors() <-chan error {
	return s.errorStream
}

// GracefulShutdown closes the Server with the grace period specified in its
// config.
func (s *Server) GracefulShutdown() error {
	s.log.Println("Shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownGracePeriod)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

func serverAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func newHandler(cfg Config) (*gin.Engine, error) {
	engine, err := newGinEngine(cfg.Env.App.Env)
	if err != nil {
		return nil, err
	}

	transferHandler := handler.NewTransferHandler(cfg.Logger, cfg.TransferService)

	engine.POST("/bulk_transfer", transferHandler.BulkTransfer)

	return engine, nil
}

func newGinEngine(env string) (*gin.Engine, error) {
	var middleware middlewareStack

	switch env {
	case "development":
		gin.SetMode(gin.DebugMode)

		middleware = globalDevelopmentMiddleware()
	case "test":
		gin.SetMode(gin.TestMode)

		middleware = globalTestMiddleware()
	default:
		return nil, EnvNotSupportedError{env: env}
	}

	engine := gin.New()
	engine.Use(middleware...)

	return engine, nil
}

// middlewareStack represents a chain of middleware in the order in which
// they'll be applied to a *gin.Engine. I.e. the first middleware in the stack
// represents the outermost layer in the HTTP call chain.
type middlewareStack []gin.HandlerFunc

// globalDevelopmentMiddleware returns the middleware stack used for all routes
// when running in development.
func globalDevelopmentMiddleware() middlewareStack {
	return middlewareStack{gin.Logger(), gin.Recovery()}
}

// globalDevelopmentMiddleware returns the middleware stack used for all routes
// when running in test.
func globalTestMiddleware() middlewareStack {
	return middlewareStack{gin.Logger(), gin.Recovery()}
}
