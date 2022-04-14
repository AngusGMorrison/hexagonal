package postgres

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// RelativeQueryDir is the path to the postgres query directory relative to the
// application root folder. This is not necessarily the same as the path
// relative to the running binary, so this should be joined to an absolute path
// representing the application root before use.
func RelativeQueryDir() string {
	return filepath.Join("internal", "adapter", "repository", "postgres", "query")
}

// QueryFilename represents the filename of an PostgreSQL query file.
type QueryFilename string

// Queries maps the name of a query file to the string query.
type Queries map[QueryFilename]string

// Load appends the named queries from queryDir into a map of filenames to their
// string contents.
//
// absQueryDir is an absolute path to a folder containing the specified query
// files.
func (q Queries) Load(absQueryDir string, filenames []QueryFilename) error {
	for _, fn := range filenames {
		path := filepath.Join(absQueryDir, string(fn))

		queryBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		q[fn] = string(queryBytes)
	}

	return nil
}
