-- name: CreateSubscription :one
INSERT INTO user_subscriptions (
    id, user_id, plan_name, status, start_date, end_date, has_used_trial, rides_left, free_parcels_quota
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: CheckHasUsedTrial :one
SELECT EXISTS (
    SELECT 1 FROM user_subscriptions
    WHERE user_id = $1 AND has_used_trial = true
);

-- name: GetActiveSubscription :one
SELECT * FROM user_subscriptions
WHERE user_id = $1 AND status = 'active' AND end_date > NOW() AND deleted_at IS NULL
ORDER BY end_date DESC LIMIT 1;

-- name: GetLatestSubscription :one
SELECT * FROM user_subscriptions
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY end_date DESC LIMIT 1;

-- name: SavePaymentToken :one
INSERT INTO payment_tokens (
    id, user_id, provider, token, masked_pan, status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetPaymentTokensByUserID :many
SELECT * FROM payment_tokens
WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL;

-- name: CreateTransaction :one
INSERT INTO transactions (
    id, user_id, amount, currency, type, status, provider, payment_token_id, external_tx_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET status = $1, external_tx_id = COALESCE($2, external_tx_id), updated_at = NOW()
WHERE id = $3 AND deleted_at IS NULL;

-- name: GetTransactionForUpdate :one
SELECT * FROM transactions
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: GetSubscriptionsForBilling :many
SELECT * FROM user_subscriptions
WHERE status IN ('active', 'past_due') AND end_date <= NOW() AND deleted_at IS NULL;

-- name: UpdateSubscriptionStatus :exec
UPDATE user_subscriptions
SET status = $1, end_date = $2, updated_at = NOW()
WHERE id = $3 AND deleted_at IS NULL;

-- name: GetPromoCodeForUpdate :one
SELECT * FROM promo_codes
WHERE code = $1 AND is_active = true AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdatePromoCodeUses :exec
UPDATE promo_codes
SET current_uses = current_uses + 1, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CheckPromoUsage :one
SELECT EXISTS (
    SELECT 1 FROM user_promo_usages
    WHERE user_id = $1 AND promo_code_id = $2 AND deleted_at IS NULL
);

-- name: RecordPromoUsage :exec
INSERT INTO user_promo_usages (
    id, user_id, promo_code_id
) VALUES (
    $1, $2, $3
);

-- name: AddFreeParcelsToPremium :exec
UPDATE user_subscriptions
SET free_parcels_quota = free_parcels_quota + $1, updated_at = NOW()
WHERE status = 'active' AND plan_name IN ('pro', 'plus') AND deleted_at IS NULL;

-- name: CancelStaleTransactions :exec
UPDATE transactions
SET status = 'cancelled', updated_at = NOW()
WHERE status = 'pending' AND created_at < NOW() - INTERVAL '30 minutes' AND deleted_at IS NULL;

-- name: DeductRideLimit :execrows
UPDATE user_subscriptions
SET rides_left = CASE WHEN rides_left = -1 THEN -1 ELSE rides_left - 1 END, updated_at = NOW()
WHERE user_id = $1 AND (rides_left > 0 OR rides_left = -1) AND status = 'active' AND end_date > NOW() AND deleted_at IS NULL;

-- name: RefundRideLimit :exec
UPDATE user_subscriptions
SET rides_left = CASE WHEN rides_left = -1 THEN -1 ELSE rides_left + 1 END, updated_at = NOW()
WHERE user_id = $1 AND status = 'active' AND end_date > NOW() AND deleted_at IS NULL;

-- name: ResetExpiredSubscriptions :execrows
UPDATE user_subscriptions
SET plan_name = 'none', status = 'expired', rides_left = 0, free_parcels_quota = 0, updated_at = NOW()
WHERE status IN ('past_due', 'canceled') AND end_date < NOW() - INTERVAL '7 days' AND deleted_at IS NULL;

-- name: DeductParcelLimit :execrows
UPDATE user_subscriptions
SET free_parcels_quota = CASE WHEN free_parcels_quota = -1 THEN -1 ELSE free_parcels_quota - 1 END, updated_at = NOW()
WHERE user_id = $1 AND (free_parcels_quota > 0 OR free_parcels_quota = -1) AND status = 'active' AND end_date > NOW() AND deleted_at IS NULL;

-- name: RefundParcelLimit :exec
UPDATE user_subscriptions
SET free_parcels_quota = CASE WHEN free_parcels_quota = -1 THEN -1 ELSE free_parcels_quota + 1 END, updated_at = NOW()
WHERE user_id = $1 AND status = 'active' AND end_date > NOW() AND deleted_at IS NULL;

-- name: UpdateActiveSubscriptionLimits :exec
UPDATE user_subscriptions
SET plan_name = $1, rides_left = $2, free_parcels_quota = $3, end_date = $4, updated_at = NOW()
WHERE id = $5 AND deleted_at IS NULL;

-- name: CancelActiveSubscription :exec
UPDATE user_subscriptions
SET plan_name = 'none', status = 'canceled', updated_at = NOW()
WHERE user_id = $1 AND status = 'active' AND end_date > NOW() AND deleted_at IS NULL;

-- name: GetTotalRevenue :one
SELECT COALESCE(SUM(amount), 0)::float8 FROM transactions WHERE status = 'success' AND type = 'subscription' AND deleted_at IS NULL;

-- name: GetRevenueChartData :many
SELECT DATE(created_at)::text as date, COALESCE(SUM(amount), 0)::float8 as amount 
FROM transactions 
WHERE status = 'success' AND type = 'subscription' AND created_at >= NOW() - INTERVAL '7 days' AND deleted_at IS NULL
GROUP BY DATE(created_at) 
ORDER BY date ASC;
