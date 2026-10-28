package postgres

import (
	"context"
	"errors"
	"testing"

	plutoauth "github.com/plutoframework/pluto-auth"
)

func TestPostgresSessionPersistenceAndRevocation(t *testing.T) {
	store := isolatedStore(t)

	sessions, err := store.SessionStore()
	if err != nil {
		t.Fatal(err)
	}

	manager, err := plutoauth.New(sessions, plutoauth.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	issued, err := manager.Start(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}

	// A second adapter instance must observe the same durable session state.
	secondStore, err := store.SessionStore()
	if err != nil {
		t.Fatal(err)
	}

	second, err := plutoauth.New(secondStore, plutoauth.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	verified, err := second.Authenticate(ctx, issued.Token)
	if err != nil || verified.SubjectID != "alice" {
		t.Fatal("session did not persist across adapters")
	}

	if err := second.Revoke(ctx, issued.Token); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.Authenticate(
		ctx,
		issued.Token,
	); !errors.Is(
		err,
		plutoauth.ErrUnauthenticated,
	) {
		t.Fatal("revocation not shared across adapters")
	}
}
