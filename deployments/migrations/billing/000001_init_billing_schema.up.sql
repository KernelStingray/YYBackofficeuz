CREATE TABLE IF NOT EXISTS user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    plan_name VARCHAR(50) NOT NULL, -- e.g. 'free', 'plus', 'pro'
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'expired', 'canceled'
    start_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_date TIMESTAMPTZ NOT NULL,
    has_used_trial BOOLEAN NOT NULL DEFAULT FALSE,
    rides_left INT NOT NULL DEFAULT 0,
    free_parcels_quota INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS payment_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    provider VARCHAR(50) NOT NULL, -- 'click', 'payme'
    token VARCHAR(255) NOT NULL,
    masked_pan VARCHAR(20) NOT NULL, -- '8600****1234'
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    amount NUMERIC NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'UZS',
    type VARCHAR(50) NOT NULL, -- 'subscription', 'ride_fee'
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'success', 'failed'
    provider VARCHAR(50) NOT NULL, -- 'click', 'payme'
    payment_token_id UUID, -- foreign key optional
    external_tx_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS promo_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_amount NUMERIC,
    discount_percent NUMERIC,
    discount_type VARCHAR(20) DEFAULT 'fixed_amount',
    max_uses INT NOT NULL DEFAULT 1,
    current_uses INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    starts_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS user_promo_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    promo_code_id UUID NOT NULL REFERENCES promo_codes(id),
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    UNIQUE(user_id, promo_code_id)
);

CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX idx_payment_tokens_user_id ON payment_tokens(user_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_external_tx_id ON transactions(external_tx_id);
CREATE INDEX idx_promo_codes_code ON promo_codes(code);
CREATE INDEX idx_user_promo_usages_user_id ON user_promo_usages(user_id);
