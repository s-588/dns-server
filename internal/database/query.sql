-- name: GetResourceRecordByID :one
SELECT id , domain , data, type_id, class_id , time_to_live ,
(SELECT type FROM types WHERE resource_records.type_id = types.id) AS type,
(SELECT class FROM classes WHERE resource_records.class_id = classes.id) AS class  
FROM resource_records
WHERE resource_records.id = $1;

-- name: GetResourceRecords :many
SELECT id , domain , data, type_id, class_id , time_to_live ,
(SELECT type FROM types WHERE resource_records.type_id = types.id) AS type,
(SELECT class FROM classes WHERE resource_records.class_id = classes.id) AS class  
FROM resource_records
WHERE domain = $1 and (SELECT types.type FROM types WHERE types.id = resource_records.type_id) = $2;

-- name: GetAllResourceRecord :many
SELECT id , domain , data, type_id, class_id , time_to_live ,
(SELECT type FROM types WHERE resource_records.type_id = types.id) AS type,
(SELECT class FROM classes WHERE resource_records.class_id = classes.id) AS class  
 FROM resource_records;

-- name: CreateResourceRecord :one
INSERT INTO resource_records (domain, data, type_id, class_id, time_to_live)
VALUES (
    $1,
    $2,
    (SELECT id FROM types WHERE type = $3),
    (SELECT id FROM classes WHERE class = $4),
    $5
)
RETURNING id;

-- name: UpdateResourceRecord :one
UPDATE resource_records
SET domain = $1,
    data = $2, 
    type_id = (SELECT id FROM types WHERE type = $3),
    class_id = (SELECT id FROM classes WHERE class = $4),
    time_to_live = $5
WHERE resource_records.id = $6
RETURNING id;

-- name: DeleteResourceRecord :exec
DELETE FROM resource_records
WHERE id = $1 ;

-- name: CreateUser :one
INSERT INTO users (login, first_name, last_name,password,role_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    (SELECT roles.id FROM roles WHERE roles.role = $5)
)
RETURNING id ;


-- name: GetUser :one
SELECT users.id, login, first_name, last_name, role, password
FROM users INNER JOIN roles ON users.role_id = roles.id
WHERE users.login = $1;

-- name: GetAllUsers :many
SELECT users.id, login, first_name, last_name, role
FROM users INNER JOIN roles ON users.role_id = roles.id;

-- name: DeleteUser :exec
DELETE FROM users
WHERE users.id = $1;

-- name: UpdateUser :exec
UPDATE users
SET login = $2, first_name = $3, last_name = $4, 
    role_id = (SELECT id FROM roles WHERE role = $5 AND role IS NOT NULL),
    password = $6
WHERE users.id = $1;
