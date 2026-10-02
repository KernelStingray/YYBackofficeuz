#!/usr/bin/env bash
# ==============================================================================
# YOLYOLAKAY ENTERPRISE ENCRYPTED DATABASE BACKUP SCRIPT
# Automated pg_dump + AES-256 GPG Encryption + S3/MinIO Offsite Sync
# ==============================================================================

set -euo pipefail

BACKUP_DIR="/var/backups/yolyolakay"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
ENCRYPT_PASS="${BACKUP_ENCRYPT_PASSWORD:-YolYolakay_Secure_2026_Backup_Key}"
S3_BUCKET="${S3_BUCKET:-s3://yolyolakay-backups/}"

mkdir -p "${BACKUP_DIR}"

DATABASES=("yolyolakay_users" "yolyolakay_rides" "yolyolakay_billing" "yolyolakay_parcels" "yolyolakay_telematics")

echo "[$(date)] Starting Enterprise Encrypted Backup..."

for DB in "${DATABASES[@]}"; do
    DUMP_FILE="${BACKUP_DIR}/${DB}_${TIMESTAMP}.sql.gz"
    ENC_FILE="${DUMP_FILE}.enc"

    echo "Dumping ${DB}..."
    docker exec -t postgres_a pg_dump -U postgres -d "${DB}" 2>/dev/null | gzip -9 > "${DUMP_FILE}" || \
    docker exec -t postgres_b pg_dump -U postgres -d "${DB}" 2>/dev/null | gzip -9 > "${DUMP_FILE}"

    echo "Encrypting ${DB} with AES-256..."
    openssl enc -aes-256-cbc -salt -in "${DUMP_FILE}" -out "${ENC_FILE}" -k "${ENCRYPT_PASS}" -pbkdf2
    rm -f "${DUMP_FILE}"

    if command -v aws &> /dev/null; then
        echo "Uploading ${ENC_FILE} to S3 (${S3_BUCKET})..."
        aws s3 cp "${ENC_FILE}" "${S3_BUCKET}${DB}/" --quiet || true
    fi
done

# Удаление локальных дампов старше 7 дней
find "${BACKUP_DIR}" -type f -name "*.enc" -mtime +7 -delete

echo "[$(date)] Enterprise Backup successfully completed!"
