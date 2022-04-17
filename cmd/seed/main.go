package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/angusgmorrison/hexagonal/envconfig"
	"github.com/angusgmorrison/hexagonal/repository/sql/database"
)

func main() {
	logger := log.New(os.Stdout, "hexagonal_seed ", log.LstdFlags)

	if err := run(); err != nil {
		logger.Fatal(err)
	}
}

func run() error {
	var (
		defaultSeedsPath = filepath.Join("fixtures", "seeds", "seeds.sql")
		seedsPath        = flag.String("path", defaultSeedsPath, "The location of the SQL seeds file to load")
	)

	flag.Parse()

	envConfig, err := envconfig.New()
	if err != nil {
		return fmt.Errorf("envconfig.New: %w", err)
	}

	db, err := database.New(envConfig.DB)
	if err != nil {
		return fmt.Errorf("postgres.NewDB: %w", err)
	}

	// TODO: Refactor
	absSeedsPath, err := filepath.Abs(*seedsPath)
	if err != nil {
		return fmt.Errorf("create absolute seeds path: %w", err)
	}

	_, err = db.LoadFile(absSeedsPath)
	if err != nil {
		return fmt.Errorf("load seeds file at %s: %w", absSeedsPath, err)
	}

	return nil
}
