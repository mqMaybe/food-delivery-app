package log

import (
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) Info(message string) {
	l.logger.Printf("[INFO] %s", message)
}

func (l *Logger) Error(message string, err error) {
	l.logger.Printf("[ERROR] %s: %v", message, err)
}

func (l *Logger) Fatal(message string, err error) {
	l.logger.Fatalf("[FATAL] %s: %v", message, err)
}
