package background

import (
	"context"
	"errors"
	"testing"

	jobs "github.com/plutoframework/pluto-jobs"
)

type recordingSink struct{ calls int }

func (s *recordingSink) ArchiveAudit(
	context.Context,
	string,
	string,
	string,
) error {
	s.calls++
	return nil
}

func TestAuditHandlerRequiresScopeAndPayload(t *testing.T) {
	tenant := "northstar"

	request, err := jobs.NewRequest(
		AuditDelivery,
		AuditReference{EventID: "event"},
		jobs.Options{TenantID: &tenant},
	)
	if err != nil {
		t.Fatal(err)
	}

	sink := &recordingSink{}
	handler := AuditHandler(sink)

	job := jobs.Job{Request: request, Attempt: 1}
	if err := handler(
		context.Background(),
		job,
	); err != nil ||
		sink.calls != 1 {
		t.Fatal(err, sink.calls)
	}

	job.TenantID = nil
	if err := handler(
		context.Background(),
		job,
	); !errors.Is(err, jobs.ErrInvalid) ||
		sink.calls != 1 {
		t.Fatal("unscoped delivery", err)
	}

	job.TenantID = &tenant

	job.Payload = []byte("{}")
	if err := handler(
		context.Background(),
		job,
	); !errors.Is(err, jobs.ErrInvalid) ||
		sink.calls != 1 {
		t.Fatal("missing event delivered", err)
	}
}

func FuzzAuditHandler(f *testing.F) {
	f.Add([]byte("{}"), "northstar")
	f.Add([]byte("{\"event_id\":\"event\"}"), "harbor")
	f.Fuzz(func(t *testing.T, payload []byte, tenant string) {
		request, err := jobs.NewRequest(
			AuditDelivery,
			AuditReference{EventID: "event"},
			jobs.Options{},
		)
		if err != nil {
			t.Fatal(err)
		}

		request.Payload, request.TenantID = payload, &tenant
		_ = AuditHandler(
			&recordingSink{},
		)(
			context.Background(),
			jobs.Job{Request: request, Attempt: 1},
		)
	})
}
