package migrate

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	// Register the postgres driver.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Register the file source.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Config is a configuration object for migrations.
type Config struct {
	// ForceVersion specifies the migration version to force-set.
	ForceVersion int

	// Rollback triggers a down migration.
	Rollback bool

	// Verbose controls the log output of the migrator's underlying
	// *migrate.Migrate.
	Verbose bool
}

// Migrate triggers a migration on the database using the specified Config.
func Migrate(databaseURL, migrationPath string, logger *log.Logger, config Config) error {
	fmt.Println(migrationPath)
	migrator, err := newMigrator(databaseURL, migrationPath, logger, config)
	if err != nil {
		return err
	}

	if shouldForce(config.ForceVersion) {
		return migrator.force()
	}

	if config.Rollback {
		return migrator.rollback()
	}

	return migrator.up()
}

// RelativeMigrationDir returns the path to the migration directory relative to
// the application root folder. This is not necessarily the same as the path
// relative to the running binary (e.g. during tests), so it should be joined
// into an absolute path before use.
func RelativeMigrationDir() string {
	return filepath.Join("internal", "repository", "sql", "migrate", "migrations")
}

type migrator struct {
	migrate *migrate.Migrate
	logger  *log.Logger
	config  Config
}

func newMigrator(databaseURL, migrationPath string, logger *log.Logger, config Config) (*migrator, error) {
	pathWithScheme := filepath.Join("file://", migrationPath)

	inner, err := migrate.New(pathWithScheme, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create new migrator: %w", err)
	}

	inner.Log = &migrateLogger{
		logger:  logger,
		verbose: config.Verbose,
	}

	migrator := migrator{
		migrate: inner,
		logger:  logger,
		config:  config,
	}

	return &migrator, nil
}

func (m *migrator) force() error {
	if err := m.migrate.Force(m.config.ForceVersion); err != nil {
		return fmt.Errorf("force migration version %d: %w", m.config.ForceVersion, err)
	}

	return nil
}

func (m *migrator) rollback() error {
	m.logger.Println("Rolling back most recent migration...")

	if err := m.migrate.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Println("Migration complete. No changes.")

			return nil
		}

		return fmt.Errorf("rollback: %w", err)
	}

	m.logger.Println("Migration complete.")

	return nil
}

func (m *migrator) up() error {
	m.logger.Println("Migrating up...")

	if err := m.migrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Println("Migration complete. No changes.")

			return nil
		}

		return fmt.Errorf("migrate up: %w", err)
	}

	m.logger.Println("Migration complete.")

	return nil
}

type migrateLogger struct {
	logger  *log.Logger
	verbose bool
}

func (ml *migrateLogger) Printf(format string, v ...interface{}) {
	ml.logger.Printf(format, v...)
}

func (ml *migrateLogger) Verbose() bool {
	return ml.verbose
}

func shouldForce(version int) bool {
	return version > 0
}
