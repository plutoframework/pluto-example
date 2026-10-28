// Command dev-fixtures is trusted local administration, never a public API.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/development"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"
)

func run() error {
	ids := map[string]*string{}
	for _, name := range []string{"alice", "bob", "carol", "dana"} {
		ids[name] = flag.String(
			name+"-id",
			"",
			"Explicit verified synthetic account ID",
		)
	}

	flag.Parse()

	settings, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := postgres.Open(ctx, settings.Database.URL)
	if err != nil {
		return fmt.Errorf("development database unavailable")
	}
	defer store.Close()

	subjects := map[string]string{}
	for name, id := range ids {
		subjects[name] = *id
	}

	if err := development.BindFixtures(
		ctx,
		store.Pool(),
		subjects,
	); err != nil {
		return fmt.Errorf(
			"development binding failed; check explicit IDs and fixture state",
		)
	}

	fmt.Println(
		"Bound DEVELOPMENT fixtures to explicit account IDs; enrollment grants no roles.",
	)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
