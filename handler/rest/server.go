package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/angusgmorrison/hexagonal/envconfig"
	"github.com/angusgmorrison/hexagonal/service"
)

// Server provides HTTP routing and handler dependencies.
type Server struct {
	config envconfig.EnvConfig
	logger *log.Logger

	// server listens for incoming HTTP requests and routes them to the correct
	// handler.
	server *http.Server

	// errorStream carries errors from the running server across process
	// boundaries. This is particularly helpful for monitoring errors from
	// multiple concurrent sources, e.g. server errors and OS interrupts.
	errorStream chan error

	// Services are the interfaces by which handlers communicate requests to
	// business logic.
	bulkTransactionService service.BulkTransactionService
}

// NewServer returns a new hexagonal server configured using the provided Config.
func NewServer(
	logger *log.Logger,
	envConfig envconfig.EnvConfig,
	transactionService service.BulkTransactionService,
) *Server {
	server := Server{
		config: envConfig,
		logger: logger,
		server: &http.Server{
			Addr:         serverAddress(envConfig.HTTP.Host, envConfig.HTTP.Port),
			ReadTimeout:  envConfig.HTTP.ReadTimeout,
			WriteTimeout: envConfig.HTTP.WriteTimeout,
		},
		errorStream:            make(chan error, 1),
		bulkTransactionService: transactionService,
	}

	server.setupRoutes()

	return &server
}

// Run starts the Server in a new goroutine, forwarding any errors to its
// errorStream. Run blocks until either a server error or an OS interrupt
// occurs. In the case of an interrupt, Run first attempts to shut down
// gracefully.
func (s *Server) Run() error {
	s.logger.Printf("Starting server at %s\n", s.server.Addr)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.errorStream <- fmt.Errorf("ListenAndServe: %w", err)
		}
	}()

	select {
	case err := <-s.errorStream:
		return fmt.Errorf("server error: %w", err)
	case sig := <-newSignalHandler():
		s.logger.Printf("Received signal %q\n", sig)

		if err := s.GracefulShutdown(); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}

		return nil
	}
}

// GracefulShutdown closes the Server with the grace period specified in its
// config.
func (s *Server) GracefulShutdown() error {
	s.logger.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.HTTP.ShutdownGracePeriod)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

// ServeHTTP calls the ServeHTTP method of the Server's underlying handler,
// passing through its ResponseWriter and Request. This is router-agnostic and
// makes the server exceptionally easy to test.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}

func serverAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func newSignalHandler() <-chan os.Signal {
	signalStream := make(chan os.Signal, 1)
	signal.Notify(signalStream, os.Interrupt, syscall.SIGTERM)

	return signalStream
}
