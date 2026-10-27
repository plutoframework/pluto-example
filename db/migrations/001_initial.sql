-- Apply once to an empty, dedicated demo database; this prototype has no migrator.
BEGIN;
CREATE TABLE organizations (
    id text PRIMARY KEY,
    name text NOT NULL
);
CREATE TABLE memberships (
    tenant_id text NOT NULL REFERENCES organizations(id),
    actor_id text NOT NULL,
    display_name text NOT NULL,
    role text NOT NULL CHECK (role IN ('member', 'approver', 'administrator')),
    PRIMARY KEY (tenant_id, actor_id)
);
CREATE TABLE purchases (
    tenant_id text NOT NULL REFERENCES organizations(id),
    id text NOT NULL,
    creator_id text NOT NULL,
    title text NOT NULL CHECK (length(btrim(title)) BETWEEN 1 AND 160),
    amount_cents bigint NOT NULL CHECK (amount_cents BETWEEN 1 AND 100000000),
    status text NOT NULL CHECK (status IN ('draft', 'submitted', 'approved', 'rejected')),
    version bigint NOT NULL CHECK (version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id, creator_id) REFERENCES memberships(tenant_id, actor_id)
);
CREATE TABLE audit_events (
    tenant_id text NOT NULL REFERENCES organizations(id),
    id text NOT NULL,
    actor_id text NOT NULL,
    action text NOT NULL,
    resource_id text NOT NULL,
    outcome text NOT NULL CHECK (outcome = 'success'),
    occurred_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id),
    FOREIGN KEY (tenant_id, actor_id) REFERENCES memberships(tenant_id, actor_id)
);
CREATE INDEX audit_events_tenant_actor_time ON audit_events (tenant_id, actor_id, occurred_at);
COMMIT;
