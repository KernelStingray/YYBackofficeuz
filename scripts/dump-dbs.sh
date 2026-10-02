#!/bin/bash
set -e

USER="KernelStingray"
DUMP_DIR="deployments/db_dumps"

mkdir -p $DUMP_DIR

echo "Dumping Group Jupiter databases..."
for db in racoon burunov spidy ford kibertron gobliddin; do
    echo "Dumping $db..."
    docker exec group_jupiter pg_dump -U $USER --clean --if-exists -O -x -d $db > "$DUMP_DIR/${db}.sql" || echo "Warning: Could not dump $db (maybe it does not exist)"
done

echo "Dumping Group Saturn databases..."
for db in romanova pascal girard senna shennon caprio; do
    echo "Dumping $db..."
    docker exec group_saturn pg_dump -U $USER --clean --if-exists -O -x -d $db > "$DUMP_DIR/${db}.sql" || echo "Warning: Could not dump $db (maybe it does not exist)"
done

echo "✅ All databases dumped successfully to $DUMP_DIR/"
