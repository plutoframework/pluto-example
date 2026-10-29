package postgres

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type countingTracer struct{ starts, ends atomic.Int64 }

func (t *countingTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	_ pgx.TraceQueryStartData,
) context.Context {
	t.starts.Add(1)

	return ctx
}

func (t *countingTracer) TraceQueryEnd(
	context.Context,
	*pgx.Conn,
	pgx.TraceQueryEndData,
) {
	t.ends.Add(1)
}

func TestPostgresPoolTracer(t *testing.T) {
	databaseURL := os.Getenv("PLUTO_EXAMPLE_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("requires isolated test PostgreSQL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tracer := &countingTracer{}

	store, err := OpenWithTracer(ctx, databaseURL, tracer)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	before := tracer.starts.Load()

	if _, err := store.Pool().Exec(ctx, "SELECT 1"); err != nil {
		t.Fatal(err)
	}

	if tracer.starts.Load() <= before ||
		tracer.starts.Load() != tracer.ends.Load() {
		t.Fatal("pool did not invoke paired query tracer callbacks")
	}
}
