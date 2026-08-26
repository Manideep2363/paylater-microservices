-- name: CreateMerchant :execresult
INSERT INTO merchants (
    name,
    email,
    phone,
    password_hash,
    commission_percentage
)
VALUES (?, ?, ?, ?, ?);

-- name: ListMerchants :many
SELECT
    merchant_id,
    name,
    phone,
    email,
    commission_percentage
FROM merchants
ORDER BY merchant_id;

-- name: GetMerchantByID :one
SELECT
    merchant_id,
    name,
    phone,
    email,
    commission_percentage
FROM merchants
WHERE merchant_id = ?;

-- name: GetMerchantByEmail :one
SELECT
    merchant_id,
    name,
    email,
    phone,
    password_hash,
    commission_percentage
FROM merchants
WHERE email = ?;

-- name: CheckMerchantEmailExists :one
SELECT EXISTS(
    SELECT 1
    FROM merchants
    WHERE email = ?
);

-- name: UpdateMerchantCommission :execresult
UPDATE merchants
SET commission_percentage = ?
WHERE merchant_id = ?;
