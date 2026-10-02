CREATE TABLE trust_scores (
    user_id UUID PRIMARY KEY,
    score NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE fraud_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- 'mock_location', 'fake_booking', etc.
    description TEXT,
    penalty_applied NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX idx_fraud_events_user_id ON fraud_events(user_id);
