CREATE TABLE tickets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    resolution TEXT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX idx_tickets_user_id ON tickets(user_id);
