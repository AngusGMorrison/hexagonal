package postgres

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type queries map[string]string

func loadQueries(appRoot string, queryFilenames []string) (queries, error) {
	absDir := absQueryDir(appRoot)

	qs := make(queries, len(queryFilenames))

	for _, filename := range queryFilenames {
		path := filepath.Join(absDir, filename)

		queryBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		qs[filename] = string(queryBytes)
	}

	return qs, nil
}

func absQueryDir(appRoot string) string {
	return filepath.Join(appRoot, "queries", "transfer")
}
