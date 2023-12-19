package logger

import (
	"go.uber.org/zap"
)

// Log is a logger instance.
// Log variable must only be changed by Initialize function.
// No-op Logger is set by default, so must be Initialized.
var Log *zap.Logger = zap.NewNop()

// Initialize configures logger with provided level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// cfg := zap.NewProductionConfig()
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl

	return nil
}

// Sync ignores err check (to avoid lint warning).
// Otherwise use logger.Log.Sync()
func Sync() {
	_ = Log.Sync()
}

// migrationLogger implements migrate.Logger interface.
type migrationLogger struct {
	logger  *zap.SugaredLogger
	verbose bool
	prefix  string
}

func NewMigrationLogger(logger *zap.Logger, verbose bool, prefix string) *migrationLogger {
	return &migrationLogger{
		logger:  logger.Sugar(),
		verbose: verbose,
		prefix:  prefix,
	}
}

// Printf is like fmt.Printf
func (ml *migrationLogger) Printf(format string, v ...interface{}) {
	ml.logger.Infof(ml.prefix+format, v...)
}

// Verbose should return true when verbose logging output is wanted
func (ml *migrationLogger) Verbose() bool {
	return ml.verbose
}
