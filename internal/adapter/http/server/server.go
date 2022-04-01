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

var (
	ErrEnvNotSupported = errors.New("environment not supported")
	ErrServerClosed    = http.ErrServerClosed
)

// Server wraps a standard library server, providing support for
// application-specific methods.
type Server struct {
	*http.Server

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
		Server: &http.Server{
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
	s.log.Printf("starting server at %s\n", s.Addr)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	s.log.Println("shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownGracePeriod)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to gracefully shut down the server: %w", err)
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
	default:
		return nil, fmt.Errorf("%w: %s", ErrEnvNotSupported, env)
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
