-- name: Organization :one
SELECT id, name FROM organizations WHERE id = $1;
