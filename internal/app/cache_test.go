package app

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/plutoframework/pluto-example/internal/api"
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

func TestPostgresAccessReflectsCurrentMembership(t *testing.T) {
	a := newLifecycleApp(t)
	cookie := a.enroll(t, "capabilities@example.test")
	w := a.request(t, "GET", "/auth/session", cookie, "", "", nil, 200)

	var session api.SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}

	a.request(t, "GET", "/api/access", cookie, "northstar", "", nil, 401)

	if _, err := a.pool.Exec(
		context.Background(),
		"INSERT INTO memberships (tenant_id,actor_id,display_name,role) VALUES ('northstar',$1,'Capabilities reader','member')",
		session.Data.ActorID,
	); err != nil {
		t.Fatal(err)
	}

	for _, role := range []string{"member", "administrator", "approver"} {
		if _, err := a.pool.Exec(
			context.Background(),
			"UPDATE memberships SET role=$1 WHERE actor_id=$2",
			role,
			session.Data.ActorID,
		); err != nil {
			t.Fatal(err)
		}

		w = a.request(
			t,
			"GET",
			"/api/access",
			cookie,
			"northstar",
			"",
			nil,
			200,
		)

		var response api.AccessResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}

		if response.Data.Role != role ||
			response.Data.ActorID != session.Data.ActorID ||
			response.Data.OrganizationID != "northstar" {
			t.Fatal("stale or crossed capability scope", response)
		}

		administrator := role == "administrator"
		if slices.Contains(
			response.Data.Permissions,
			"memberships:manage",
		) != administrator ||
			(len(response.Data.Roles) > 0) != administrator {
			t.Fatal("capability display drift", response)
		}
	}

	a.request(t, "GET", "/api/access", cookie, "harbor", "", nil, 401)
}
