//

package memory

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"

	"github.com/lazygpt/lazygpt/plugin/log"
)

type Logger struct {
	*log.Logger
}

var _ badger.Logger = (*Logger)(nil)

// NewLogger creates a new Logger instance. It supports the `badger.Logger`
// interface and can be used to set the logger for the badger database.
func NewLogger(logger *log.Logger) *Logger {
	return &Logger{
		Logger: logger,
	}
}

// Errorf implements the `badger.Logger` interface.
func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
}

// Warningf implements the `badger.Logger` interface.
func (logger *Logger) Warningf(format string, args ...interface{}) {
	logger.Warn(fmt.Sprintf(format, args...))
}

// Infof implements the `badger.Logger` interface.
func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Debugf implements the `badger.Logger` interface.
func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}
