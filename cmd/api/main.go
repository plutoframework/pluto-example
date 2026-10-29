package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plutoframework/pluto-example/internal/app"
	"github.com/plutoframework/pluto-example/internal/config"

	"github.com/plutoframework/pluto-core/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	logger, closeLogger, err := app.NewLogger(settings.Logging)
	if err != nil {
		return err
	}
	defer closeLogger()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	telemetryConfig, err := config.LoadTelemetry(
		"pluto-example-api",
		settings.Cache.Environment,
	)
	if err != nil {
		return err
	}

	runtime, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		return err
	}

	logger = slog.New(runtime.Handler(logger.Handler()))

	defer func() {
		shutdown, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := runtime.Shutdown(shutdown); err != nil {
			logger.Error("telemetry shutdown failed")
		}
	}()

	startup, cancelStartup := context.WithTimeout(ctx, 10*time.Second)

	application, err := app.NewInstrumented(startup, settings, logger, runtime)

	cancelStartup()

	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}

	defer application.Close()

	server := &http.Server{
		Addr:              settings.Server.Address,
		Handler:           application.Handler,
		ReadHeaderTimeout: settings.Server.ReadTimeout,
		ReadTimeout:       settings.Server.ReadTimeout,
		WriteTimeout:      settings.Server.WriteTimeout,
		IdleTimeout:       settings.Server.IdleTimeout,
		MaxHeaderBytes:    16 * 1024,
	}

	stopped := make(chan error, 1)
	go func() { stopped <- server.ListenAndServe() }()

	logger.Info(
		"local demo listening",
		"address",
		settings.Server.Address,
		"storage",
		settings.Database.Mode,
	)

	select {
	case err := <-stopped:
		if err != http.ErrServerClosed {
			return err
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdown); err != nil {
			return err
		}
	}

	return nil
}
