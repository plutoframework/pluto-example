package app

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPostgresOrganizationCacheKeepsMembershipLive(t *testing.T) {
	a := newLifecycleApp(t)
	cookie := a.enroll(t, "cache-reader@example.test")
	w := a.request(t, "GET", "/auth/session", cookie, "", "", nil, 200)

	var session struct {
		Data struct {
			ActorID string `json:"actor_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}

	a.request(t, "GET", "/api/organization", cookie, "northstar", "", nil, 401)

	// Explicit trusted test provisioning: enrollment never grants membership.
	_, err := a.pool.Exec(
		context.Background(),
		"INSERT INTO memberships (tenant_id,actor_id,display_name,role) VALUES ('northstar',$1,'Cache reader','member'),('harbor',$1,'Cache reader','member')",
		session.Data.ActorID,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, tenant := range []string{"northstar", "northstar", "harbor"} {
		w := a.request(
			t,
			"GET",
			"/api/organization",
			cookie,
			tenant,
			"",
			nil,
			200,
		)

		var profile struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(
			w.Body.Bytes(),
			&profile,
		); err != nil ||
			profile.Data.ID != tenant {
			t.Fatal("profile crossed tenant", err)
		}
	}

	_, err = a.pool.Exec(
		context.Background(),
		"DELETE FROM memberships WHERE tenant_id='northstar' AND actor_id=$1",
		session.Data.ActorID,
	)
	if err != nil {
		t.Fatal(err)
	}

	a.request(t, "GET", "/api/organization", cookie, "northstar", "", nil, 401)
	a.request(t, "GET", "/api/organization", cookie, "harbor", "", nil, 200)
}
