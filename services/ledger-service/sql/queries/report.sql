-- name: GetMerchantCommissionSummary :many
SELECT
    merchant_id,
    CAST(
        COALESCE(SUM(commission_amount), 0.00)
        AS DECIMAL(10,2)
    ) AS total_commission
FROM transactions
GROUP BY merchant_id
ORDER BY total_commission DESC;
