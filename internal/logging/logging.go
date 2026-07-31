package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var LogLevel = &slog.LevelVar{} //nolint:gochecknoglobals

func SetLevel(level string) error {
	_level, err := parseLogLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level (%w)", err)
	}

	LogLevel.Set(_level)

	return nil
}

func Level() string {
	return strings.ToLower(LogLevel.Level().String())
}

func Init(level string) {
	if err := SetLevel(level); err != nil {
		slog.Error("failed to set init log level", "level", level)
		LogLevel.Set(slog.LevelInfo)
	}

	opts := &slog.HandlerOptions{
		Level: LogLevel,
	}
	newLogger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(newLogger)
}

func parseLogLevel(levelStr string) (slog.Level, error) {
	var level slog.Level

	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		return slog.Level(0), fmt.Errorf("invalid log level: %s (%w)", levelStr, err)
	}

	return level, nil
}
