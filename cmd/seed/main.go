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
	logger := log.New(os.Stdout, "hexagonal_seed ", log.LstdFlags)

	if err := run(); err != nil {
		logger.Fatal(err)
	}
}

func run() error {
	defaultSeedsPath := filepath.Join("fixtures", "seeds", "seeds.sql")
	seedsPath := flag.String("path", defaultSeedsPath, "The location of the SQL seeds file to load")
	flag.Parse()

	env, err := envconfig.New()
	if err != nil {
		return fmt.Errorf("envconfig.New: %w", err)
	}

	pg, err := postgres.New(env.DB)
	if err != nil {
		return fmt.Errorf("postgres.New: %w", err)
	}

	absSeedsPath, err := filepath.Abs(*seedsPath)
	if err != nil {
		return fmt.Errorf("create absolute seeds path: %w", err)
	}

	_, err = pg.LoadFile(absSeedsPath)
	if err != nil {
		return fmt.Errorf("load seeds file at %s: %w", absSeedsPath, err)
	}

	return nil
}
