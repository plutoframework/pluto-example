-- name: GetMembership :one
SELECT * FROM memberships WHERE tenant_id = $1 AND actor_id = $2;

-- name: ListMemberships :many
SELECT * FROM memberships WHERE tenant_id = $1 ORDER BY actor_id;

-- name: ChangeRole :execrows
UPDATE memberships SET role = $3 WHERE tenant_id = $1 AND actor_id = $2;

-- name: PageMemberships :many
SELECT * FROM memberships WHERE tenant_id = $1
AND actor_id > sqlc.arg(after_id)::text
ORDER BY actor_id LIMIT sqlc.arg(fetch_limit)::int;
