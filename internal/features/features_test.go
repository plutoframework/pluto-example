package features

import (
	"testing"

	"github.com/plutoframework/pluto-example/internal/domain"
)

func TestSnapshotPresenceAndExposure(t *testing.T) {
	actor := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}

	for _, test := range []struct {
		name      string
		overrides map[Key]bool
		want      bool
	}{
		{
			name: "absent fallback",
			want: false,
		},
		{
			name:      "explicit false",
			overrides: map[Key]bool{PurchaseExport: false},
			want:      false,
		},
		{
			name:      "explicit true",
			overrides: map[Key]bool{PurchaseExport: true},
			want:      true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			snapshot := NewSnapshot(test.overrides)
			if snapshot.Enabled(PurchaseExport, actor) != test.want {
				t.Fatal("wrong effective value")
			}

			visible := Visible(snapshot, actor)
			if len(visible) != 2 || visible["purchase_review"] ||
				visible["purchase_export"] != test.want {
				t.Fatalf("bad exposure: %+v", visible)
			}

			if !snapshot.Enabled("internal_diagnostics", actor) {
				t.Fatal("internal fallback should be true")
			}

			if snapshot.Enabled("unknown", actor) ||
				snapshot.Enabled(PurchaseExport, domain.Principal{}) {
				t.Fatal("unknown or unverified evaluation enabled")
			}
		})
	}

	input := map[Key]bool{PurchaseExport: true}
	snapshot := NewSnapshot(input)

	input[PurchaseExport] = false
	if !snapshot.Enabled(PurchaseExport, actor) {
		t.Fatal("snapshot aliases mutable input")
	}
}
