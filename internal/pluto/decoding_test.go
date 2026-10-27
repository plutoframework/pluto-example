package pluto

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/plutoframework/pluto-example/internal/api"
)

func FuzzDecodePurchase(f *testing.F) {
	f.Add(`{"title":"Safety goggles","amount_cents":2500}`)
	f.Add(`{"title":"x","amount_cents":1,"organization_id":"harbor"}`)
	f.Add(`{} {}`)
	f.Add(`{"amount_cents":1.5}`)
	f.Add(`null`)
	f.Add(`{"Title":"x","AMOUNT_CENTS":1}`)
	f.Fuzz(func(t *testing.T, body string) {
		request := httptest.NewRequest("POST", "/", strings.NewReader(body))
		writer := httptest.NewRecorder()

		var input api.CreatePurchaseRequest

		err := DecodeJSON(writer, request, &input, 1024)
		if err != nil {
			return
		}

		// Accepted input must be one valid bounded JSON value with no unknown keys.
		if len(body) > 1024 || !json.Valid([]byte(body)) {
			t.Fatal("accepted oversized or invalid JSON")
		}

		var fields map[string]json.RawMessage
		if err := json.Unmarshal(
			[]byte(body),
			&fields,
		); err != nil ||
			fields == nil {
			t.Fatalf("accepted non-object: %v", err)
		}

		for key := range fields {
			if key != "title" && key != "amount_cents" {
				t.Fatalf("accepted unknown field %q", key)
			}
		}

		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		var roundtrip api.CreatePurchaseRequest
		if err := json.Unmarshal(
			encoded,
			&roundtrip,
		); err != nil ||
			roundtrip != input {
			t.Fatal("accepted model does not round-trip")
		}
	})
}
