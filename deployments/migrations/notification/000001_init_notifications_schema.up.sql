CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS chats_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL,
    receiver_id UUID NOT NULL,
    ride_id UUID,
    reply_to_id UUID REFERENCES chats_messages(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'sent',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_chats_messages_sender_receiver ON chats_messages (sender_id, receiver_id);
CREATE INDEX idx_chats_messages_ride ON chats_messages (ride_id);

CREATE TABLE IF NOT EXISTS device_tokens (
    user_id UUID PRIMARY KEY,
    fcm_token VARCHAR(255) NOT NULL,
    platform VARCHAR(50),
    language VARCHAR(10) DEFAULT 'ru',
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    ride_id UUID DEFAULT NULL,
    booking_id UUID DEFAULT NULL,
    parcel_id UUID DEFAULT NULL,
    data JSONB DEFAULT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC);
