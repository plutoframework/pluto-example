package operations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
)

type purchaseFake struct {
	Repository      // Unexpected methods panic: each unit test declares only its collaborators.
	current         domain.Purchase
	readTenant      string
	writes          int
	saved           domain.Purchase
	expectedVersion int64
	event           domain.AuditEvent
	saveError       error
}

func (f *purchaseFake) Purchase(
	_ context.Context,
	tenant, id string,
) (domain.Purchase, error) {
	f.readTenant = tenant
	if tenant != f.current.TenantID || id != f.current.ID {
		return domain.Purchase{}, domain.ErrNotFound
	}

	return f.current, nil
}

func (f *purchaseFake) SavePurchase(
	_ context.Context,
	purchase domain.Purchase,
	version int64,
	event domain.AuditEvent,
) error {
	f.writes++
	f.saved = purchase
	f.expectedVersion = version
	f.event = event

	return f.saveError
}

func fixture(status domain.Status) domain.Purchase {
	return domain.Purchase{
		ID:          "request-1",
		TenantID:    "northstar",
		CreatorID:   "alice",
		Title:       "Microscope service",
		AmountCents: 45000,
		Status:      status,
		Version:     4,
	}
}

func principal(actor string, role domain.Role) domain.Principal {
	return domain.Principal{
		ActorID:  actor,
		TenantID: "northstar",
		Role:     role,
	}
}

func testService(repository Repository) *Service {
	service := New(repository, features.NewSnapshot(nil))
	service.now = func() time.Time { return time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC) }
	service.newID = func() string { return "event-or-request-id" }

	return service
}

func TestDecisionPolicies(t *testing.T) {
	tests := []struct {
		name     string
		actor    string
		role     domain.Role
		status   domain.Status
		decision domain.Status
		want     error
	}{
		{
			name:     "approve submitted",
			actor:    "bob",
			role:     domain.Approver,
			status:   domain.Submitted,
			decision: domain.Approved,
		},
		{
			name:     "reject submitted",
			actor:    "carol",
			role:     domain.Administrator,
			status:   domain.Submitted,
			decision: domain.Rejected,
		},
		{
			name:     "member cannot decide",
			actor:    "dana",
			role:     domain.Member,
			status:   domain.Submitted,
			decision: domain.Approved,
			want:     domain.ErrForbidden,
		},
		{
			name:     "no self approval",
			actor:    "alice",
			role:     domain.Approver,
			status:   domain.Submitted,
			decision: domain.Approved,
			want:     domain.ErrForbidden,
		},
		{
			name:     "no self rejection",
			actor:    "alice",
			role:     domain.Administrator,
			status:   domain.Submitted,
			decision: domain.Rejected,
			want:     domain.ErrForbidden,
		},
		{
			name:     "draft cannot be approved",
			actor:    "bob",
			role:     domain.Approver,
			status:   domain.Draft,
			decision: domain.Approved,
			want:     domain.ErrConflict,
		},
		{
			name:     "terminal cannot be decided",
			actor:    "bob",
			role:     domain.Approver,
			status:   domain.Approved,
			decision: domain.Rejected,
			want:     domain.ErrConflict,
		},
		{
			name:     "unknown decision",
			actor:    "bob",
			role:     domain.Approver,
			status:   domain.Submitted,
			decision: "pending",
			want:     domain.ErrInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &purchaseFake{current: fixture(test.status)}

			result, err := testService(
				fake,
			).Decide(context.Background(), principal(test.actor, test.role), "request-1", test.decision)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v; want %v", err, test.want)
			}

			if test.want != nil {
				if fake.writes != 0 {
					t.Fatal("denied decision wrote state")
				}

				return
			}

			if result.Status != test.decision || result.Version != 5 ||
				fake.expectedVersion != 4 {
				t.Fatalf(
					"bad transition: %+v, expected version %d",
					result,
					fake.expectedVersion,
				)
			}

			if fake.event.TenantID != "northstar" ||
				fake.event.ActorID != test.actor ||
				fake.event.ResourceID != "request-1" ||
				fake.event.Outcome != "success" {
				t.Fatalf("bad audit event: %+v", fake.event)
			}

			if fake.event.Action != "purchase."+string(test.decision) {
				t.Fatal("wrong audit action")
			}
		})
	}
}

func TestSubmitPolicies(t *testing.T) {
	for _, test := range []struct {
		name, actor string
		status      domain.Status
		want        error
	}{
		{
			"creator submits draft",
			"alice",
			domain.Draft,
			nil,
		},
		{
			"other member cannot submit",
			"bob",
			domain.Draft,
			domain.ErrNotFound,
		},
		{
			"cannot submit twice",
			"alice",
			domain.Submitted,
			domain.ErrConflict,
		},
		{
			"cannot resubmit rejected",
			"alice",
			domain.Rejected,
			domain.ErrConflict,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := &purchaseFake{current: fixture(test.status)}

			value, err := testService(
				fake,
			).Submit(context.Background(), principal(test.actor, domain.Member), "request-1")
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v; want %v", err, test.want)
			}

			if test.want != nil && fake.writes != 0 {
				t.Fatal("denied submission wrote state")
			}

			if test.want == nil &&
				(value.Status != domain.Submitted || fake.expectedVersion != 4 || fake.event.Action != "purchase.submitted") {
				t.Fatalf("bad submission: %+v", value)
			}
		})
	}
}

func TestCrossTenantGetAndDecision(t *testing.T) {
	fake := &purchaseFake{current: fixture(domain.Submitted)}
	service := testService(fake)

	actor := domain.Principal{
		ActorID:  "dana",
		TenantID: "harbor",
		Role:     domain.Administrator,
	}
	if _, err := service.Get(
		context.Background(),
		actor,
		"request-1",
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatalf("cross tenant get: %v", err)
	}

	if _, err := service.Decide(
		context.Background(),
		actor,
		"request-1",
		domain.Approved,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatalf("cross tenant decide: %v", err)
	}

	if fake.readTenant != "harbor" || fake.writes != 0 {
		t.Fatal("tenant scope was not retained")
	}
}

func TestCreateValidationAndAudit(t *testing.T) {
	for _, test := range []struct {
		name, title string
		cents       int64
		want        error
	}{
		{
			"valid",
			"  Team safety equipment  ",
			12500,
			nil,
		},
		{
			"empty",
			" ",
			12500,
			domain.ErrInvalid,
		},
		{
			"negative",
			"Equipment",
			-1,
			domain.ErrInvalid,
		},
		{
			"zero",
			"Equipment",
			0,
			domain.ErrInvalid,
		},
		{
			"too expensive",
			"Equipment",
			100000001,
			domain.ErrInvalid,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := &purchaseFake{}

			result, err := testService(
				fake,
			).Create(context.Background(), principal("alice", domain.Member), test.title, test.cents)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v; want %v", err, test.want)
			}

			if test.want != nil && fake.writes != 0 {
				t.Fatal("invalid creation wrote state")
			}

			if test.want == nil &&
				(result.Title != "Team safety equipment" || result.Status != domain.Draft || fake.event.Action != "purchase.created" || fake.expectedVersion != 0) {
				t.Fatalf("bad creation: %+v", result)
			}
		})
	}
}

func TestConflictIsPropagated(t *testing.T) {
	fake := &purchaseFake{
		current:   fixture(domain.Draft),
		saveError: domain.ErrConflict,
	}

	_, err := testService(
		fake,
	).Submit(context.Background(), principal("alice", domain.Member), "request-1")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("conflict lost: %v", err)
	}
}

func TestExportPermissionAndFeatureAreIndependent(t *testing.T) {
	// No repository methods are implemented: denied calls must stop before data access.
	fake := &purchaseFake{}

	service := testService(fake)
	if _, err := service.Export(
		context.Background(),
		principal("bob", domain.Approver),
	); !errors.Is(
		err,
		domain.ErrFeatureDisabled,
	) {
		t.Fatalf("disabled export: %v", err)
	}

	service.features = features.NewSnapshot(
		map[features.Key]bool{features.PurchaseExport: true},
	)
	if _, err := service.Export(
		context.Background(),
		principal("alice", domain.Member),
	); !errors.Is(
		err,
		domain.ErrForbidden,
	) {
		t.Fatalf("flag must not grant permission: %v", err)
	}

	if _, err := service.Create(
		context.Background(),
		domain.Principal{},
		"x",
		1,
	); !errors.Is(
		err,
		domain.ErrUnauthenticated,
	) {
		t.Fatalf("empty principal accepted: %v", err)
	}
}

func FuzzDecisionPolicy(f *testing.F) {
	f.Add("submitted", "approver", "approved", false)
	f.Add("submitted", "administrator", "approved", true)
	f.Add("draft", "member", "rejected", false)
	f.Fuzz(func(t *testing.T, status, role, decision string, self bool) {
		fake := &purchaseFake{current: fixture(domain.Status(status))}

		actor := "bob"
		if self {
			actor = "alice"
		}

		_, err := testService(
			fake,
		).Decide(context.Background(), principal(actor, domain.Role(role)), "request-1", domain.Status(decision))

		allowed := (role == "approver" || role == "administrator") && !self &&
			status == "submitted" &&
			(decision == "approved" || decision == "rejected")
		if allowed {
			if err != nil || fake.writes != 1 || fake.saved.Version != 5 ||
				string(fake.saved.Status) != decision {
				t.Fatalf(
					"valid decision rejected or corrupted: %v %+v",
					err,
					fake.saved,
				)
			}
		} else if err == nil || fake.writes != 0 {
			t.Fatalf(
				"invalid decision wrote state: err=%v writes=%d",
				err,
				fake.writes,
			)
		}
	})
}
