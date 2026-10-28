-- name: GetPurchase :one
SELECT * FROM purchases WHERE tenant_id = $1 AND id = $2;

-- name: ListPurchases :many
SELECT * FROM purchases WHERE tenant_id = $1 ORDER BY id;

-- name: CreatePurchase :exec
INSERT INTO purchases (tenant_id, id, creator_id, title, amount_cents, status, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdatePurchase :execrows
UPDATE purchases SET status = $3, version = $4, updated_at = $5
WHERE tenant_id = $1 AND id = $2 AND version = $6;

-- name: PagePurchases :many
SELECT * FROM purchases
WHERE tenant_id = $1
AND (sqlc.arg(actor_filter)::text = '' OR creator_id = sqlc.arg(actor_filter))
AND id > sqlc.arg(after_id)::text
ORDER BY id LIMIT sqlc.arg(fetch_limit)::int;
