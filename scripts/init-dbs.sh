#!/bin/bash
set -e

USER="KernelStingray"
MAX_RETRIES=30

wait_for_db() {
  local container=$1
  local count=0
  echo "Waiting for $container to be ready..."
  until docker exec "$container" pg_isready -U "$USER" 2>/dev/null; do
    count=$((count + 1))
    if [ "$count" -ge "$MAX_RETRIES" ]; then
      echo "❌ Error: $container failed to become ready after $MAX_RETRIES attempts."
      exit 1
    fi
    sleep 2
  done
  echo "✅ $container is ready!"
}

if docker ps --format '{{.Names}}' | grep -q "^group_jupiter$"; then
  wait_for_db "group_jupiter"
  echo "Creating databases in group_jupiter (Cluster A - Uzbekistan)..."
  docker exec group_jupiter psql -U $USER -d group_jupiter -tc "SELECT 1 FROM pg_database WHERE datname = 'racoon'" | grep -q 1 || docker exec group_jupiter psql -U $USER -d group_jupiter -c "CREATE DATABASE racoon;"
  docker exec group_jupiter psql -U $USER -d group_jupiter -tc "SELECT 1 FROM pg_database WHERE datname = 'burunov'" | grep -q 1 || docker exec group_jupiter psql -U $USER -d group_jupiter -c "CREATE DATABASE burunov;"
  docker exec group_jupiter psql -U $USER -d group_jupiter -tc "SELECT 1 FROM pg_database WHERE datname = 'spidy'" | grep -q 1 || docker exec group_jupiter psql -U $USER -d group_jupiter -c "CREATE DATABASE spidy;"
  docker exec group_jupiter psql -U $USER -d group_jupiter -tc "SELECT 1 FROM pg_database WHERE datname = 'ford'" | grep -q 1 || docker exec group_jupiter psql -U $USER -d group_jupiter -c "CREATE DATABASE ford;"
  docker exec group_jupiter psql -U $USER -d group_jupiter -tc "SELECT 1 FROM pg_database WHERE datname = 'kibertron'" | grep -q 1 || docker exec group_jupiter psql -U $USER -d group_jupiter -c "CREATE DATABASE kibertron;"
  docker exec group_jupiter psql -U $USER -d group_jupiter -tc "SELECT 1 FROM pg_database WHERE datname = 'gobliddin'" | grep -q 1 || docker exec group_jupiter psql -U $USER -d group_jupiter -c "CREATE DATABASE gobliddin;"
fi

if docker ps --format '{{.Names}}' | grep -q "^group_saturn$"; then
  wait_for_db "group_saturn"
  echo "Creating databases in group_saturn (Cluster B - Europe)..."
  docker exec group_saturn psql -U $USER -d group_saturn -tc "SELECT 1 FROM pg_database WHERE datname = 'romanova'" | grep -q 1 || docker exec group_saturn psql -U $USER -d group_saturn -c "CREATE DATABASE romanova;"
  docker exec group_saturn psql -U $USER -d group_saturn -tc "SELECT 1 FROM pg_database WHERE datname = 'pascal'" | grep -q 1 || docker exec group_saturn psql -U $USER -d group_saturn -c "CREATE DATABASE pascal;"
  docker exec group_saturn psql -U $USER -d group_saturn -tc "SELECT 1 FROM pg_database WHERE datname = 'girard'" | grep -q 1 || docker exec group_saturn psql -U $USER -d group_saturn -c "CREATE DATABASE girard;"
  docker exec group_saturn psql -U $USER -d group_saturn -tc "SELECT 1 FROM pg_database WHERE datname = 'senna'" | grep -q 1 || docker exec group_saturn psql -U $USER -d group_saturn -c "CREATE DATABASE senna;"
  docker exec group_saturn psql -U $USER -d group_saturn -tc "SELECT 1 FROM pg_database WHERE datname = 'shennon'" | grep -q 1 || docker exec group_saturn psql -U $USER -d group_saturn -c "CREATE DATABASE shennon;"
  docker exec group_saturn psql -U $USER -d group_saturn -tc "SELECT 1 FROM pg_database WHERE datname = 'caprio'" | grep -q 1 || docker exec group_saturn psql -U $USER -d group_saturn -c "CREATE DATABASE caprio;"
fi

echo "Database initialization completed."

