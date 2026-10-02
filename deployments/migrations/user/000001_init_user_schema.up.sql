CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    gender VARCHAR(20),
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    role VARCHAR(20) NOT NULL DEFAULT 'passenger', -- passenger, driver
    kyc_status VARCHAR(20) DEFAULT 'none', -- none, pending, approved, rejected
    kyc_rejection_reason VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, blocked
    locale VARCHAR(10) DEFAULT 'uz',
    avatar_url VARCHAR(500),
    rating DOUBLE PRECISION NOT NULL DEFAULT 5.0,
    trusted_devices TEXT[] DEFAULT '{}',
    rating_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE cars (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    gov_number VARCHAR(20) UNIQUE NOT NULL,
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    color VARCHAR(50),
    year INT DEFAULT 0,
    tech_passport_photo VARCHAR(500),
    driver_license_photo VARCHAR(500),
    tech_passport_front VARCHAR(500),
    tech_passport_back VARCHAR(500),
    driver_license_front VARCHAR(500),
    driver_license_back VARCHAR(500),
    car_front_photo VARCHAR(500),
    car_back_photo VARCHAR(500),
    car_side_photo VARCHAR(500),
    moderation_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE user_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL, -- passport, driver_license, vehicle_registration, selfie
    file_url VARCHAR(500) NOT NULL,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ride_id UUID NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_cars_user_id ON cars(user_id);

-- Add test user for Google Play closed testing
INSERT INTO users (
    id, phone_number, first_name, last_name, role, kyc_status, status, locale
) VALUES (
    '00000000-0000-0000-0000-000000000001', '+998991234567', 'Test', 'User', 'driver', 'approved', 'active', 'ru'
) ON CONFLICT (id) DO NOTHING;
CREATE INDEX idx_user_documents_user_id ON user_documents(user_id);
CREATE INDEX idx_reviews_users ON reviews(reviewer_id, reviewee_id);

CREATE TABLE user_blocks (
    blocker_id UUID NOT NULL,
    blocked_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (blocker_id, blocked_id)
);

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id UUID NOT NULL,
    reported_id UUID NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
