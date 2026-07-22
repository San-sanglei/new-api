#!/bin/sh
# ============================================================================
# Docker 容器内定时备份脚本（由 db-backup 容器的 cron 调用）
#
# 此脚本运行在 postgres:15 容器内，通过 pg_dump 备份到 /backups 目录
# 保留最近 BACKUP_RETAIN_DAYS 天的备份（默认 7 天）
#
# 环境变量（由 docker-compose.yml 注入）：
#   PGHOST / PGPORT / PGUSER / PGDATABASE / PGPASSWORD
#   BACKUP_RETAIN_DAYS - 备份保留天数，默认 7
# ============================================================================

set -eu

BACKUP_DIR="/backups"
RETAIN_DAYS="${BACKUP_RETAIN_DAYS:-7}"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BACKUP_FILE="${BACKUP_DIR}/backup_${TIMESTAMP}.sql"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] 开始数据库备份: ${PGDATABASE}@${PGHOST}:${PGPORT}"

# 执行逻辑备份
pg_dump \
    -h "$PGHOST" \
    -p "$PGPORT" \
    -U "$PGUSER" \
    -d "$PGDATABASE" \
    --format=plain \
    --no-owner \
    --no-privileges \
    -f "$BACKUP_FILE"

# 验证备份文件
if [ ! -f "$BACKUP_FILE" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] 备份文件不存在: $BACKUP_FILE"
    exit 1
fi

FILE_SIZE=$(stat -c%s "$BACKUP_FILE" 2>/dev/null || stat -f%z "$BACKUP_FILE" 2>/dev/null)
if [ "$FILE_SIZE" -lt 100 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] 备份文件过小 (${FILE_SIZE} 字节)，可能备份失败"
    exit 1
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] 备份完成: $BACKUP_FILE (${FILE_SIZE} 字节)"

# 清理过期备份
echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] 清理超过 ${RETAIN_DAYS} 天的备份..."
find "$BACKUP_DIR" -name "backup_*.sql" -type f -mtime +${RETAIN_DAYS} -print -delete | while read -r deleted; do
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] 已删除过期备份: $(basename "$deleted")"
done

echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] 备份流程完成"
