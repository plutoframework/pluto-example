-- name: AppendAudit :exec
INSERT INTO audit_events (tenant_id, id, actor_id, action, resource_id, outcome, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAudit :many
SELECT * FROM audit_events
WHERE tenant_id = $1 AND (sqlc.arg(actor_filter)::text = '' OR actor_id = sqlc.arg(actor_filter))
ORDER BY occurred_at, id;
