-- name: ArchiveAudit :one
INSERT INTO audit_archive (
    tenant_id, event_id, job_id, actor_id, action, resource_id, outcome, occurred_at
)
SELECT events.tenant_id, events.id, sqlc.arg(job_id)::text, events.actor_id,
       events.action, events.resource_id, events.outcome, events.occurred_at
FROM audit_events AS events
WHERE events.tenant_id = sqlc.arg(tenant_id)::text
AND events.id = sqlc.arg(event_id)::text
ON CONFLICT (tenant_id, event_id) DO UPDATE
SET job_id = audit_archive.job_id
RETURNING event_id;
