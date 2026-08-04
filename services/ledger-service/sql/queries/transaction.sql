-- name: CreateTransaction :execresult
INSERT INTO transactions (
    user_id,
    merchant_id,
    amount,
    commission_percentage,
    commission_amount
)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: ListTransactions :many
SELECT
    transaction_id,
    user_id,
    merchant_id,
    amount,
    commission_percentage,
    commission_amount,
    created_at
FROM transactions
ORDER BY transaction_id DESC;

-- name: GetTransactionByID :one
SELECT
    transaction_id,
    user_id,
    merchant_id,
    amount,
    commission_percentage,
    commission_amount,
    created_at
FROM transactions
WHERE transaction_id = ?;

-- name: ListUserTransactions :many
SELECT
    transaction_id,
    user_id,
    merchant_id,
    amount,
    commission_percentage,
    commission_amount,
    created_at
FROM transactions
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: ListMerchantTransactions :many
SELECT
    transaction_id,
    user_id,
    merchant_id,
    amount,
    commission_percentage,
    commission_amount,
    created_at
FROM transactions
WHERE merchant_id = ?
ORDER BY created_at DESC;
