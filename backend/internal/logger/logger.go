package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

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

func Debug(format string, args ...interface{}) {
	logf(LevelDebug, "DEBUG", format, args...)
}

func Info(format string, args ...interface{}) {
	logf(LevelInfo, "INFO", format, args...)
}

func Error(format string, args ...interface{}) {
	logf(LevelError, "ERROR", format, args...)
}

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

func logf(level Level, prefix, format string, args ...interface{}) {
	if !enabled {
		return
	}
	if level > currentLevel {
		return
	}
	stdLogger.Printf(fmt.Sprintf("[%s] %s", prefix, format), args...)
}
