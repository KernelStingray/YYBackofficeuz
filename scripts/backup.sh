#!/bin/bash
set -e

# Required Environment Variables:
# PGHOST, PGUSER, PGPASSWORD, PGDATABASE
# S3_BUCKET (e.g. s3://my-bucket/backups/)
# AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION (and optionally AWS_ENDPOINT_URL)
# BACKUP_ENCRYPTION_KEY (optional)

DATE=$(date +"%Y-%m-%d_%H-%M-%S")
FILENAME="backup_${PGDATABASE}_${DATE}.sql.gz"
TMP_DIR="/tmp/pg_backup"

mkdir -p "$TMP_DIR"
cd "$TMP_DIR"

echo "[$(date)] Starting backup for database: $PGDATABASE"

# 1. pg_dump and gzip
# Uses .pgpass or PGPASSWORD environment variable
if [ "$PGDATABASE" = "all" ]; then
    echo "[$(date)] Running pg_dumpall..."
    pg_dumpall -h "$PGHOST" -U "$PGUSER" | gzip > "$FILENAME"
else
    pg_dump -h "$PGHOST" -U "$PGUSER" "$PGDATABASE" | gzip > "$FILENAME"
fi

# 2. Encrypt if BACKUP_ENCRYPTION_KEY is provided
UPLOAD_FILE="$FILENAME"
if [ -n "$BACKUP_ENCRYPTION_KEY" ]; then
    echo "[$(date)] Encrypting backup archive..."
    # Warning: Passing password via -k in CLI can be insecure on shared hosts.
    # We use it here inside an isolated Docker container.
    openssl enc -aes-256-cbc -salt -pbkdf2 -in "$FILENAME" -out "${FILENAME}.enc" -k "$BACKUP_ENCRYPTION_KEY"
    UPLOAD_FILE="${FILENAME}.enc"
    rm -f "$FILENAME" # remove unencrypted archive
fi

# 3. Upload to S3
echo "[$(date)] Uploading $UPLOAD_FILE to $S3_BUCKET..."
if [ -n "$AWS_ENDPOINT_URL" ]; then
    aws s3 cp "$UPLOAD_FILE" "${S3_BUCKET%/}/${UPLOAD_FILE}" --endpoint-url "$AWS_ENDPOINT_URL"
else
    aws s3 cp "$UPLOAD_FILE" "${S3_BUCKET%/}/${UPLOAD_FILE}"
fi

# 4. Cleanup
rm -f "$UPLOAD_FILE"
echo "[$(date)] Backup completed successfully."
