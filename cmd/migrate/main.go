package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
)

const migrationsPath = "file://migrations"

var (
	forceVersion int
	rollback     bool
	verbose      bool
)

func main() {
	logger := log.New(os.Stdout, "hexagonal_migrate ", log.LstdFlags)

	flag.IntVar(&forceVersion, "force", 0, "force migration to the specified version")
	flag.BoolVar(&rollback, "rollback", false, "roll back the last migration (default: false)")
	flag.BoolVar(&verbose, "verbose", true, "toggle verbose logging (default: true)")
	flag.Parse()

	env, err := envconfig.New()
	if err != nil {
		logger.Panic(err)
	}

	migrator, err := migrate.New(migrationsPath, env.DB.URL())
	if err != nil {
		logger.Panic(err)
	}

	migrator.Log = migrateLogger{logger}

	if forceVersion > 0 {
		if err = migrator.Force(forceVersion); err != nil {
			logger.Printf("failed to force migration version %d: %s\n", forceVersion, err)

			return
		}
	}

	if rollback {
		if err = migrateDown(migrator, logger); err != nil {
			logger.Printf("migrateDown failed: %s\n", err)
		}

		return
	}

	if err = migrateUp(migrator, logger); err != nil {
		logger.Printf("migrateUp failed: %s\n", err)

		return
	}
}

func migrateUp(migrator *migrate.Migrate, logger *log.Logger) error {
	logger.Println("Starting UP migrations...")

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Println("Migration complete. No changes.")

			return nil
		}

		return fmt.Errorf("migrator.Up failed: %w", err)
	}

	logger.Println("Migration complete.")

	return nil
}

func migrateDown(migrator *migrate.Migrate, logger *log.Logger) error {
	logger.Println("Rolling back most recent migration...")

	if err := migrator.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Println("Migration complete. No changes.")

			return nil
		}

		return fmt.Errorf("migrator.Steps(-1) failed: %w", err)
	}

	logger.Println("Migration complete.")

	return nil
}

type migrateLogger struct {
	*log.Logger
}

func (ml migrateLogger) Verbose() bool {
	return verbose
}
