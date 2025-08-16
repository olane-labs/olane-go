package core

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// DefaultLogger implements the Logger interface
type DefaultLogger struct {
	name     string
	level    LogLevel
	logger   *log.Logger
	colorize bool
}

// NewLogger creates a new logger with the given name
func NewLogger(name string) Logger {
	logger := &DefaultLogger{
		name:     name,
		level:    LogLevelInfo,
		logger:   log.New(os.Stdout, "", 0),
		colorize: true,
	}

	// Check if DEBUG environment variable is set
	if debugEnv := os.Getenv("DEBUG"); debugEnv != "" {
		if strings.Contains(debugEnv, "*") || strings.Contains(debugEnv, name) {
			logger.level = LogLevelDebug
		}
	}

	return logger
}

// SetLevel sets the logging level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// shouldLog checks if a message at the given level should be logged
func (l *DefaultLogger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// formatMessage formats a log message with timestamp, level, and name
func (l *DefaultLogger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	
	if l.colorize {
		return l.colorizeMessage(timestamp, level, l.name, message)
	}
	
	return fmt.Sprintf("%s [%s] %s: %s", timestamp, level.String(), l.name, message)
}

// colorizeMessage adds ANSI color codes to the log message
func (l *DefaultLogger) colorizeMessage(timestamp string, level LogLevel, name, message string) string {
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorGray   = "\033[90m"
		colorCyan   = "\033[36m"
	)
	
	var levelColor string
	switch level {
	case LogLevelDebug:
		levelColor = colorGray
	case LogLevelInfo:
		levelColor = colorBlue
	case LogLevelWarn:
		levelColor = colorYellow
	case LogLevelError:
		levelColor = colorRed
	}
	
	return fmt.Sprintf("%s%s%s [%s%s%s] %s%s%s: %s",
		colorGray, timestamp, colorReset,
		levelColor, level.String(), colorReset,
		colorCyan, name, colorReset,
		message)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(args ...interface{}) {
	if l.shouldLog(LogLevelDebug) {
		message := l.formatMessage(LogLevelDebug, "%s", fmt.Sprint(args...))
		l.logger.Println(message)
	}
}

// Info logs an info message
func (l *DefaultLogger) Info(args ...interface{}) {
	if l.shouldLog(LogLevelInfo) {
		message := l.formatMessage(LogLevelInfo, "%s", fmt.Sprint(args...))
		l.logger.Println(message)
	}
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(args ...interface{}) {
	if l.shouldLog(LogLevelWarn) {
		message := l.formatMessage(LogLevelWarn, "%s", fmt.Sprint(args...))
		l.logger.Println(message)
	}
}

// Error logs an error message
func (l *DefaultLogger) Error(args ...interface{}) {
	if l.shouldLog(LogLevelError) {
		message := l.formatMessage(LogLevelError, "%s", fmt.Sprint(args...))
		l.logger.Println(message)
	}
}

// Debugf logs a formatted debug message
func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	if l.shouldLog(LogLevelDebug) {
		message := l.formatMessage(LogLevelDebug, format, args...)
		l.logger.Println(message)
	}
}

// Infof logs a formatted info message
func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	if l.shouldLog(LogLevelInfo) {
		message := l.formatMessage(LogLevelInfo, format, args...)
		l.logger.Println(message)
	}
}

// Warnf logs a formatted warning message
func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	if l.shouldLog(LogLevelWarn) {
		message := l.formatMessage(LogLevelWarn, format, args...)
		l.logger.Println(message)
	}
}

// Errorf logs a formatted error message
func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	if l.shouldLog(LogLevelError) {
		message := l.formatMessage(LogLevelError, format, args...)
		l.logger.Println(message)
	}
}

// NoOpLogger is a logger that does nothing (useful for testing)
type NoOpLogger struct{}

// NewNoOpLogger creates a new no-op logger
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Debug(args ...interface{})                   {}
func (l *NoOpLogger) Info(args ...interface{})                    {}
func (l *NoOpLogger) Warn(args ...interface{})                    {}
func (l *NoOpLogger) Error(args ...interface{})                   {}
func (l *NoOpLogger) Debugf(format string, args ...interface{})   {}
func (l *NoOpLogger) Infof(format string, args ...interface{})    {}
func (l *NoOpLogger) Warnf(format string, args ...interface{})    {}
func (l *NoOpLogger) Errorf(format string, args ...interface{})   {}
