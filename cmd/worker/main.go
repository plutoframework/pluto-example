package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plutoframework/pluto-example/internal/app"
	"github.com/plutoframework/pluto-example/internal/background"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"

	"github.com/plutoframework/pluto-core/telemetry"
	jobpostgres "github.com/plutoframework/pluto-jobs/stores/postgres"
)

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
		"pluto-example-worker",
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

	startup, cancel := context.WithTimeout(ctx, 10*time.Second)

	store, err := postgres.OpenWithTracer(
		startup,
		settings.Database.URL,
		runtime.PGXTracer(),
	)

	cancel()

	if err != nil {
		return errors.New("worker database unavailable")
	}

	defer store.Close()

	outbox, err := jobpostgres.New(store.Pool())
	if err != nil {
		return err
	}

	worker, err := background.NewInstrumentedWorker(
		outbox,
		store,
		settings.Jobs,
		runtime.JobHandler,
	)
	if err != nil {
		return err
	}

	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return errors.New("worker stopped; inspect sanitized job diagnostics")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
