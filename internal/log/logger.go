package log

import (
	"log"
	"os"
)

// Logger структура для логирования
type Logger struct {
	logger *log.Logger
}

// NewLogger создаёт новый экземпляр Logger
func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info логирует информационное сообщение
func (l *Logger) Info(message string) {
	l.logger.Printf("[INFO] %s", message)
}

// Error логирует сообщение об ошибке
func (l *Logger) Error(message string, err error) {
	l.logger.Printf("[ERROR] %s: %v", message, err)
}

// Fatal логирует фатальную ошибку и завершает программу
func (l *Logger) Fatal(message string, err error) {
	l.logger.Fatalf("[FATAL] %s: %v", message, err)
}
