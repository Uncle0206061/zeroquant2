#!/bin/bash
# PostgreSQL 每日备份脚本
# 执行时间：凌晨2点（crontab: 0 2 * * * /app/scripts/backup/pg_backup.sh）
# 保留天数：7天

set -euo pipefail

# ============ 配置 ============
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-biz}"
BACKUP_DIR="/backup/postgresql"
RETENTION_DAYS=7

# ============ 执行备份 ============
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p "${BACKUP_DIR}"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting backup: ${DB_NAME}"

# pg_dump + gzip 压缩
PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --format=plain \
    --no-owner \
    --no-privileges \
    | gzip > "${BACKUP_FILE}"

if [ $? -eq 0 ]; then
    SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup completed: ${BACKUP_FILE} (${SIZE})"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup FAILED!" >&2
    exit 1
fi

# ============ 清理过期备份 ============
DELETED=$(find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql.gz" -mtime +${RETENTION_DAYS} -delete -print | wc -l)
if [ "${DELETED}" -gt 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaned ${DELETED} old backup(s) (>${RETENTION_DAYS} days)"
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup job finished"
