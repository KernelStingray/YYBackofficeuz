#!/bin/bash
set -e

# Uses golang-migrate docker image to run migrations
MIGRATE_IMG="migrate/migrate:v4.16.2"

# Cluster A (Uzbekistan / group_jupiter)
if docker ps --format '{{.Names}}' | grep -q "^group_jupiter$"; then
  echo "Running migrations for group_jupiter (Cluster A - Uzbekistan)..."
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/user -database "postgresql://KernelStingray:KeRSt!n90%248!@group_jupiter:5432/racoon?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/notification -database "postgresql://KernelStingray:KeRSt!n90%248!@group_jupiter:5432/burunov?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/helpdesk -database "postgresql://KernelStingray:KeRSt!n90%248!@group_jupiter:5432/spidy?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/antifraud -database "postgresql://KernelStingray:KeRSt!n90%248!@group_jupiter:5432/ford?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/backoffice -database "postgresql://KernelStingray:KeRSt!n90%248!@group_jupiter:5432/kibertron?sslmode=disable" up || true
fi

# Cluster B (Europe / group_saturn)
if docker ps --format '{{.Names}}' | grep -q "^group_saturn$"; then
  echo "Running migrations for group_saturn (Cluster B - Europe)..."
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/ride -database "postgresql://KernelStingray:KeRSt!n90%248!@group_saturn:5432/senna?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/billing -database "postgresql://KernelStingray:KeRSt!n90%248!@group_saturn:5432/girard?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/analytics -database "postgresql://KernelStingray:KeRSt!n90%248!@group_saturn:5432/shennon?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/parcel -database "postgresql://KernelStingray:KeRSt!n90%248!@group_saturn:5432/romanova?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/telematics -database "postgresql://KernelStingray:KeRSt!n90%248!@group_saturn:5432/pascal?sslmode=disable" up || true
  docker run --rm -v "$(pwd)/deployments":/deployments --network vpn_network $MIGRATE_IMG -path /deployments/migrations/rating -database "postgresql://KernelStingray:KeRSt!n90%248!@group_saturn:5432/caprio?sslmode=disable" up || true
fi

echo "Migrations completed successfully."
