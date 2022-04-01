package migrate

import "log"

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
