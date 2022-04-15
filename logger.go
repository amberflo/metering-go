package metering

import (
	"log"
	"os"
)

type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}

type AmberfloDefaultLogger struct {
	logger *log.Logger
}

func NewAmberfloDefaultLogger() *AmberfloDefaultLogger {
	return &AmberfloDefaultLogger{logger: log.New(os.Stdout, "amberflo.io ", log.LstdFlags)}
}

func (l *AmberfloDefaultLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *AmberfloDefaultLogger) Logf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}
