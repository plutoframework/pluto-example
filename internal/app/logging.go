package app

import (
	"io"
	"log/slog"
	"os"

	"github.com/plutoframework/pluto-example/internal/config"
)

// NewLogger fans JSON records out to configured writers; diagnostic logging is
// independent from the transactionally stored business audit events.
func NewLogger(settings config.LoggingConfig) (*slog.Logger, func(), error) {
	destinations := []io.Writer{}
	closeFile := func() {}

	if settings.Stdout {
		destinations = append(destinations, os.Stdout)
	}

	if settings.File != "" {
		file, err := os.OpenFile(
			settings.File,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0o600,
		)
		if err != nil {
			return nil, nil, err
		}

		destinations = append(destinations, file)
		closeFile = func() { _ = file.Close() }
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(settings.Level)); err != nil {
		closeFile()

		return nil, nil, err
	}

	handler := slog.NewJSONHandler(
		io.MultiWriter(destinations...),
		&slog.HandlerOptions{Level: level},
	)

	return slog.New(handler), closeFile, nil
}
