package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Level defines how verbose logs should be (error-only, info, or debug).
type Level int

const (
	LevelError Level = iota
	LevelInfo
	LevelDebug
)

var (
	currentLevel Level = LevelInfo
	enabled            = true
	stdLogger          = log.New(os.Stdout, "", log.LstdFlags)
)

// Configure sets the active log level and whether logging is enabled.
func Configure(level string, isEnabled bool) {
	enabled = isEnabled
	currentLevel = parseLevel(level)
}

// Debug logs detailed technical information helpful during development.
func Debug(format string, args ...interface{}) {
	logf(LevelDebug, "DEBUG", format, args...)
}

// Info logs important normal runtime events.
func Info(format string, args ...interface{}) {
	logf(LevelInfo, "INFO", format, args...)
}

// Error logs failures that need attention.
func Error(format string, args ...interface{}) {
	logf(LevelError, "ERROR", format, args...)
}

// parseLevel turns a text setting (from env/config) into an internal log level.
func parseLevel(level string) Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// logf is the shared implementation that enforces enabled flag and minimum log level.
func logf(level Level, prefix, format string, args ...interface{}) {
	if !enabled {
		return
	}
	if level > currentLevel {
		return
	}
	stdLogger.Printf(fmt.Sprintf("[%s] %s", prefix, format), args...)
}
