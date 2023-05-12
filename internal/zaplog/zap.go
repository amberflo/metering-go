package zaplog

import (
	"strings"

	"go.uber.org/zap"
)

// Logger provides the methods needed for a logger in the metering-go package.
// This wraps the zap logger and in some cases formats data to ensure it prints
// out nicely.
type ZapLogger struct {
	logger *zap.SugaredLogger
}

// NewLogger returns a new Logger from a zap logger
func NewZapLogger(l *zap.SugaredLogger) *ZapLogger {
	return &ZapLogger{logger: l}
}

func (l *ZapLogger) Log(v ...interface{})                 { l.logger.Info(v...) }                   //nolint:revive
func (l *ZapLogger) Logf(format string, v ...interface{}) { l.logger.Infof(l.clean(format), v...) } //nolint:revive

func (l *ZapLogger) clean(str string) string {
	str = strings.TrimPrefix(str, "metering-go: ")
	return strings.TrimSuffix(str, "\n")
}
