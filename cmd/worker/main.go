package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plutoframework/pluto-example/internal/background"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"

	jobpostgres "github.com/plutoframework/pluto-jobs/stores/postgres"
)

func run() error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	startup, cancel := context.WithTimeout(ctx, 10*time.Second)

	store, err := postgres.Open(startup, settings.Database.URL)

	cancel()

	if err != nil {
		return errors.New("worker database unavailable")
	}

	defer store.Close()

	outbox, err := jobpostgres.New(store.Pool())
	if err != nil {
		return err
	}

	worker, err := background.NewWorker(outbox, store, settings.Jobs)
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
