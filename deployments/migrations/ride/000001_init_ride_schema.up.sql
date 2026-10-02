-- Устанавливаем расширение PostGIS
CREATE EXTENSION IF NOT EXISTS postgis;

-- Таблица для хранения информации о поездках
CREATE TABLE IF NOT EXISTS rides (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,   -- Привязка к User Service
    vehicle_id UUID NOT NULL,  -- Привязка к User Service
    origin GEOMETRY(Point, 4326) NOT NULL,      -- Точка отправления (Широта, Долгота)
    destination GEOMETRY(Point, 4326) NOT NULL, -- Точка прибытия
    origin_address VARCHAR NOT NULL,
    destination_address VARCHAR NOT NULL,
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    total_seats INT NOT NULL CHECK (total_seats > 0),
    available_seats INT NOT NULL CHECK (available_seats >= 0),
    price_per_seat DECIMAL(10, 2) NOT NULL,
    gender_restriction VARCHAR(20) DEFAULT NULL,
    status VARCHAR NOT NULL DEFAULT 'created', -- 'created', 'active', 'completed', 'canceled'
    is_negotiable BOOLEAN NOT NULL DEFAULT FALSE,
    has_fair_price_badge BOOLEAN NOT NULL DEFAULT FALSE,
    has_ac BOOLEAN NOT NULL DEFAULT FALSE,
    has_luggage BOOLEAN NOT NULL DEFAULT FALSE,
    takes_parcels BOOLEAN NOT NULL DEFAULT FALSE,
    no_smoking BOOLEAN NOT NULL DEFAULT FALSE,
    no_pets BOOLEAN NOT NULL DEFAULT FALSE,
    driving_speed VARCHAR(50) DEFAULT 'normal',
    max_two_back BOOLEAN NOT NULL DEFAULT FALSE,
    pickup_point VARCHAR(255) NOT NULL DEFAULT '',
    pickup_lat DOUBLE PRECISION DEFAULT NULL,
    pickup_lon DOUBLE PRECISION DEFAULT NULL,
    music_taste VARCHAR(100) DEFAULT NULL,
    allows_city_detour BOOLEAN DEFAULT false,
    city_detour_fee NUMERIC(10, 2) DEFAULT 0,
    vehicle_data JSONB,
    route_polygon GEOMETRY(Polygon, 4326),
    cancel_reason TEXT,
    cancelled_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Пространственные (Spatial) индексы для быстрого поиска ближайших поездок
CREATE INDEX IF NOT EXISTS idx_rides_origin ON rides USING GIST (origin);
CREATE INDEX IF NOT EXISTS idx_rides_destination ON rides USING GIST (destination);
CREATE INDEX IF NOT EXISTS idx_rides_route_polygon ON rides USING GIST (route_polygon);

-- Индекс для фильтрации активных поездок
CREATE INDEX IF NOT EXISTS idx_rides_status ON rides(status);

-- Таблица для хранения остановок маршрута
CREATE TABLE IF NOT EXISTS ride_stops (
    id UUID PRIMARY KEY,
    ride_id UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326) NOT NULL,
    address VARCHAR,
    order_index INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_ride_stops_location ON ride_stops USING GIST (location);

-- Таблица для хранения бронирований
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    ride_id UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    passenger_id UUID NOT NULL, -- Привязка к User Service
    seats_booked INT NOT NULL CHECK (seats_booked > 0),
    total_price DECIMAL(10, 2) NOT NULL,
    is_full_car_buyout BOOLEAN NOT NULL DEFAULT FALSE,
    offered_price DECIMAL(10, 2) NULL,
    status VARCHAR NOT NULL DEFAULT 'pending', -- 'pending', 'confirmed', 'canceled'
    pickup_preference VARCHAR(255) DEFAULT 'wait_at_route',
    dropoff_type VARCHAR(20) DEFAULT 'highway',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Индексы для бронирований
CREATE INDEX IF NOT EXISTS idx_bookings_ride_id ON bookings(ride_id);
CREATE INDEX IF NOT EXISTS idx_bookings_passenger_id ON bookings(passenger_id);

-- Подписки на маршруты
CREATE TABLE IF NOT EXISTS route_subscriptions (
    id UUID PRIMARY KEY,
    passenger_id UUID NOT NULL,
    origin_lat DOUBLE PRECISION NOT NULL,
    origin_lon DOUBLE PRECISION NOT NULL,
    dest_lat DOUBLE PRECISION NOT NULL,
    dest_lon DOUBLE PRECISION NOT NULL,
    departure_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_route_subs_passenger ON route_subscriptions(passenger_id);
CREATE INDEX IF NOT EXISTS idx_route_subs_active ON route_subscriptions(is_active) WHERE is_active = true;

-- Отправленные напоминания (защита от повторов)
CREATE TABLE IF NOT EXISTS sent_reminders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id UUID NOT NULL,
    user_id UUID NOT NULL,
    kind VARCHAR(32) NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_ride_user_kind UNIQUE (ride_id, user_id, kind)
);

