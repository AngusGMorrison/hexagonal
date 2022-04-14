package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
)

func main() {
	logger := log.New(os.Stdout, "hexagonal_migrate ", log.LstdFlags)

	if err := run(logger); err != nil {
		logger.Fatal(err)
	}
}

func run(logger *log.Logger) error {
	var (
		forceVersion = flag.Int("force", 0, "force migration to the specified version")
		rollback     = flag.Bool("rollback", false, "roll back the last migration (default: false)")
		verbose      = flag.Bool("verbose", true, "toggle verbose logging (default: true)")
	)

	flag.Parse()

	env, err := envconfig.New()
	if err != nil {
		return fmt.Errorf("envconfig.New: %w", err)
	}

	config := postgres.MigrateConfig{
		ForceVersion: *forceVersion,
		Rollback:     *rollback,
		Verbose:      *verbose,
	}

	migrationPath := filepath.Join(env.App.Root, postgres.RelativeMigrationDir())

	if err := postgres.Migrate(env.DB.URL(), migrationPath, logger, config); err != nil {
		return fmt.Errorf("trigger migration: %w", err)
	}

	return nil
}
