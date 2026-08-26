-- name: CreateUser :execresult
INSERT INTO users (
    name,
    email,
    password
)
VALUES (
    ?,
    ?,
    ?
);

-- name: ListUsers :many
SELECT
    user_id,
    name,
    email,
    credit_limit,
    current_due
FROM users
ORDER BY user_id;

-- name: GetUserByID :one
SELECT
    user_id,
    name,
    email,
    credit_limit,
    current_due
FROM users
WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT
    user_id,
    name,
    email,
    password,
    credit_limit,
    current_due
FROM users
WHERE email = ?;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1
    FROM users
    WHERE email = ?
);

-- name: GetUserByIDForUpdate :one
SELECT
    user_id,
    name,
    email,
    credit_limit,
    current_due
FROM users
WHERE user_id = ?
FOR UPDATE;

-- name: IncreaseUserDue :execresult
UPDATE users
SET current_due = ?
WHERE user_id = ?;

-- name: DecreaseUserDue :execresult
UPDATE users
SET current_due = ?
WHERE user_id = ?;
