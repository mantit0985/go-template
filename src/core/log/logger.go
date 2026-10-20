package log

import (
	"log/slog"
	"os"
	"strings"
)

var (
	// programLevel allows runtime changes to the logging level.
	programLevel = &slog.LevelVar{}
)

// PII keys that should be redacted from logs.
var piiKeys = map[string]struct{}{
	"password": {},
	"token":    {},
	"secret":   {},
	"email":    {},
	"apikey":   {},
	"auth":     {},
}

// redactAttr replaces the value of PII attributes with a mask.
func redactAttr(groups []string, a slog.Attr) slog.Attr {
	if _, ok := piiKeys[strings.ToLower(a.Key)]; ok {
		return slog.String(a.Key, "[REDACTED]")
	}
	return a
}

// InitLogger initializes the global slog logger with JSON output,
// dynamic level control, and PII redaction.
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: programLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// First, handle PII redaction
			a = redactAttr(groups, a)

			// Next, integrate OpenTelemetry trace IDs if present in the context
			// Note: This is a simplification. In a full implementation, a custom
			// slog.Handler would extract this from the context of each log call.
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	// Default level to Info.
	programLevel.Set(slog.LevelInfo)

	slog.Info("Logger initialized", "level", programLevel.Level())
}

// SetLevel allows changing the log level at runtime.
func SetLevel(level slog.Level) {
	programLevel.Set(level)
}

// GetLevel returns the current logging level.
func GetLevel() slog.Level {
	return programLevel.Level()
}
