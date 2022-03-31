package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/http/server"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres/transferrepo"
	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain"
)

func main() {
	logger := log.New(os.Stdout, "hexagonal_migrate ", log.LstdFlags)

	// Load environment variables.
	env, err := envconfig.GetEnv()
	if err != nil {
		logger.Panic(err)
	}

	// Set up the server's IO dependencies.
	db, err := postgres.New(env.DB)
	if err != nil {
		logger.Panic(err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			logger.Panic(fmt.Errorf("failed to close db: %w", err))
		}
	}()

	transferRepo := transferrepo.New(db, logger)
	transferService := transferdomain.NewService(transferRepo)
	serverConfig := server.Config{
		Env:             env,
		Logger:          logger,
		TransferService: transferService,
	}

	// Create the server, injecting dependencies.
	svr, err := server.New(serverConfig)
	if err != nil {
		logger.Panic(err)
	}

	svr.Run()

	// Monitor the running program for server errors and interrupts.
	if err = supervise(svr, logger); err != nil {
		logger.Panic(err)
	}
}

// supervise monitors the running program, shutting down if the server fails or
// the process is interrupted by the user.
func supervise(server *server.Server, logger *log.Logger) error {
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-server.Errors():
		return fmt.Errorf("server error: %w", err)
	case sig := <-interrupts:
		logger.Printf("Received signal %q. Shutting down gracefully...\n", sig)

		err := server.GracefulShutdown()
		if err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
	}

	return nil
}
