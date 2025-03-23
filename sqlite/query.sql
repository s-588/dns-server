-- name: GetResourceRecord :many
SELECT * FROM resourceRecords
WHERE domain = ?;

-- name: GetResourceRecordRecursive :many
SELECT * FROM resourceRecords
WHERE result = ?;

-- name: CreateResourceRecord :one
INSERT INTO resourceRecords (
	id, domain, result, type, class, ttl
) values (
?,?,?,?,?,?
)
RETURNING *;

-- name: DeleteResourceRecord :exec
DELETE FROM resourceRecords
WHERE id = ?;
