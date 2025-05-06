-- name: GetResourceRecord :many
SELECT * FROM resrecords
WHERE domain = ?;

-- name: GetResourceRecordRecursive :many
SELECT * FROM resrecords ;

-- name: CreateResourceRecord :one
INSERT INTO resrecords (domain, data, typeID, classID, ttl)
VALUES (
    ?,
    ?,
    (SELECT ID FROM types WHERE type = ?),
    (SELECT ID FROM classes WHERE class = ?),
    ?
)
RETURNING *;

-- name: GetTypeName :one
SELECT type FROM types WHERE ID = ?;

-- name: GetClassName :one
SELECT class FROM classes WHERE ID = ?;

-- name: DeleteResourceRecord :one
DELETE FROM resrecords
WHERE id = ?
RETURNING *;
