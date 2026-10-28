// Package development contains trusted, opt-in fixture administration. It is
// never called by HTTP enrollment, sign-in, email matching or startup.
package development

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/plutoframework/pluto-example/internal/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BindFixtures accepts explicit account IDs from the trusted developer command.
// Fixed synthetic email checks prevent accidentally binding real accounts.
func BindFixtures(
	ctx context.Context,
	pool *pgxpool.Pool,
	subjects map[string]string,
) error {
	if len(subjects) != 4 {
		return errors.New("four explicit fixture account IDs required")
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		rollback, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
		defer cancel()

		_ = tx.Rollback(rollback)
	}()

	for _, name := range []string{"alice", "bob", "carol", "dana"} {
		subject := subjects[name]
		if subject == "" {
			return errors.New("fixture account ID missing")
		}

		var eligible bool

		err := tx.QueryRow(ctx, `SELECT verified AND NOT disabled
            FROM pluto_auth_accounts WHERE id = $1 AND email = $2 FOR UPDATE`,
			subject, name+"@example.test").Scan(&eligible)
		if err != nil || !eligible {
			return errors.New(
				"fixture account must be verified, enabled and synthetic",
			)
		}
	}

	for _, member := range auth.DemoMemberships() {
		subject := subjects[member.ActorID]

		result, err := tx.Exec(ctx, `INSERT INTO memberships
            (tenant_id, actor_id, display_name, role) VALUES ($1, $2, $3, $4)
            ON CONFLICT (tenant_id, actor_id) DO NOTHING`,
			member.TenantID, subject, member.DisplayName, string(member.Role))
		if err != nil {
			return err
		}

		if result.RowsAffected() == 0 {
			continue
		}

		_, err = tx.Exec(ctx, `INSERT INTO audit_events
            (tenant_id, id, actor_id, action, resource_id, outcome, occurred_at)
            VALUES ($1, $2, $3, 'development.fixture_bound', $3, 'success', $4)`,
			member.TenantID, rand.Text(), subject, time.Now().UTC())
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
