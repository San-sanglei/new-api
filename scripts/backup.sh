#!/usr/bin/env bash
# ============================================================================
# 数据库自动备份脚本
#
# 支持数据库类型：SQLite / PostgreSQL / MySQL
#
# 用法：
#   ./scripts/backup.sh                          # 自动检测数据库类型
#   DB_TYPE=postgres ./scripts/backup.sh         # 指定 PostgreSQL
#   DB_TYPE=mysql ./scripts/backup.sh            # 指定 MySQL
#   DB_TYPE=sqlite ./scripts/backup.sh           # 指定 SQLite
#
# 环境变量（优先级高于自动检测）：
#   DB_TYPE          - 数据库类型：sqlite / postgres / mysql
#   BACKUP_DIR       - 备份目录，默认 ./backups
#   BACKUP_RETAIN_DAYS - 备份保留天数，默认 7
#   SQLITE_PATH      - SQLite 数据库文件路径，默认 one-api.db
#   SQL_DSN          - PostgreSQL/MySQL 连接字符串（与 docker-compose 中一致）
#   POSTGRES_PASSWORD - PostgreSQL 密码（若 SQL_DSN 未含密码）
#   MYSQL_PASSWORD    - MySQL 密码（若 SQL_DSN 未含密码）
#
# 自动化（crontab）：
#   0 3 * * * cd /path/to/new-api && ./scripts/backup.sh >> ./logs/backup.log 2>&1
#   每天凌晨 3 点执行备份
# ============================================================================

set -euo pipefail

# ---------- 颜色输出 ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info()  { echo -e "${GREEN}[INFO]${NC}  $(date '+%Y-%m-%d %H:%M:%S') $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $(date '+%Y-%m-%d %H:%M:%S') $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*" >&2; }

# ---------- 配置 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

BACKUP_DIR="${BACKUP_DIR:-$PROJECT_DIR/backups}"
BACKUP_RETAIN_DAYS="${BACKUP_RETAIN_DAYS:-7}"
SQLITE_PATH="${SQLITE_PATH:-$PROJECT_DIR/one-api.db}"

TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BACKUP_FILE="$BACKUP_DIR/backup_${TIMESTAMP}"

mkdir -p "$BACKUP_DIR"

# ---------- 自动检测数据库类型 ----------
detect_db_type() {
    if [[ -n "${DB_TYPE:-}" ]]; then
        echo "$DB_TYPE"
        return
    fi
    local dsn="${SQL_DSN:-}"
    if [[ "$dsn" == postgres://* || "$dsn" == postgresql://* ]]; then
        echo "postgres"
    elif [[ "$dsn" == mysql://* || "$dsn" == *tcp(* ]]; then
        echo "mysql"
    elif [[ -f "$SQLITE_PATH" ]]; then
        echo "sqlite"
    elif [[ -z "$dsn" ]]; then
        echo "sqlite"
    else
        echo "unknown"
    fi
}

# ---------- 解析 PostgreSQL DSN ----------
# 输入: postgresql://user:pass@host:port/dbname
# 输出: PG_USER PG_PASS PG_HOST PG_PORT PG_DB
parse_pg_dsn() {
    local dsn="$1"
    # 去掉协议头
    local rest="${dsn#postgresql://}"
    rest="${rest#postgres://}"
    # 提取 user:pass
    local auth="${rest%%@*}"
    PG_USER="${auth%%:*}"
    PG_PASS="${auth#*:}"
    [[ "$PG_PASS" == "$auth" ]] && PG_PASS=""
    # 提取 host:port/dbname
    local hostpart="${rest#*@}"
    PG_HOST="${hostpart%%:*}"
    local portdb="${hostpart#*:}"
    PG_PORT="${portdb%%/*}"
    PG_DB="${portdb#*/}"
    # 如果端口为空，用默认值
    [[ -z "$PG_PORT" ]] && PG_PORT="5432"
    # 如果 PG_PASS 为空，尝试从 POSTGRES_PASSWORD 读取
    [[ -z "$PG_PASS" ]] && PG_PASS="${POSTGRES_PASSWORD:-}"
}

# ---------- 解析 MySQL DSN ----------
# 输入: root:pass@tcp(host:port)/dbname
parse_mysql_dsn() {
    local dsn="$1"
    local auth="${dsn%%@*}"
    MYSQL_USER="${auth%%:*}"
    MYSQL_PASS="${auth#*:}"
    [[ "$MYSQL_PASS" == "$auth" ]] && MYSQL_PASS=""
    local hostpart="${dsn#*@tcp(}"
    hostpart="${hostpart%%)*}"
    MYSQL_HOST="${hostpart%%:*}"
    MYSQL_PORT="${hostpart#*:}"
    local dbpart="${dsn#*/}"
    MYSQL_DB="${dbpart%%\?*}"
    # 去掉参数
    MYSQL_DB="${MYSQL_DB%%\?*}"
    [[ -z "$MYSQL_PORT" ]] && MYSQL_PORT="3306"
    [[ -z "$MYSQL_PASS" ]] && MYSQL_PASS="${MYSQL_PASSWORD:-}"
}

# ---------- 备份 SQLite ----------
backup_sqlite() {
    local db_file="$SQLITE_PATH"
    if [[ ! -f "$db_file" ]]; then
        log_error "SQLite 数据库文件不存在: $db_file"
        log_error "请通过 SQLITE_PATH 环境变量指定正确路径"
        exit 1
    fi
    BACKUP_FILE="${BACKUP_FILE}.db"
    log_info "正在备份 SQLite 数据库: $db_file"

    # 使用 sqlite3 的 .backup 命令进行一致性备份（安全处理并发写入）
    if command -v sqlite3 &>/dev/null; then
        sqlite3 "$db_file" ".backup '$BACKUP_FILE'"
    else
        # 无 sqlite3 命令时使用 cp（配合 WAL checkpoint）
        log_warn "未找到 sqlite3 命令，使用 cp 备份（可能不一致）"
        cp "$db_file" "$BACKUP_FILE"
        # 同时备份 WAL 和 SHM 文件（如果存在）
        [[ -f "${db_file}-wal" ]] && cp "${db_file}-wal" "${BACKUP_FILE}-wal"
        [[ -f "${db_file}-shm" ]] && cp "${db_file}-shm" "${BACKUP_FILE}-shm"
    fi

    local size=$(du -h "$BACKUP_FILE" | cut -f1)
    log_info "SQLite 备份完成: $BACKUP_FILE ($size)"
}

# ---------- 备份 PostgreSQL ----------
backup_postgres() {
    local dsn="${SQL_DSN:-}"
    if [[ -z "$dsn" ]]; then
        log_error "SQL_DSN 环境变量未设置，无法备份 PostgreSQL"
        exit 1
    fi
    parse_pg_dsn "$dsn"
    BACKUP_FILE="${BACKUP_FILE}.sql"
    log_info "正在备份 PostgreSQL 数据库: ${PG_USER}@${PG_HOST}:${PG_PORT}/${PG_DB}"

    # 优先使用 pg_dump
    if command -v pg_dump &>/dev/null; then
        PGPASSWORD="$PG_PASS" pg_dump \
            -h "$PG_HOST" \
            -p "$PG_PORT" \
            -U "$PG_USER" \
            -d "$PG_DB" \
            --format=plain \
            --no-owner \
            --no-privileges \
            -f "$BACKUP_FILE"
    else
        # Docker 环境下通过容器执行
        log_info "未找到 pg_dump，尝试通过 Docker 容器执行..."
        docker exec postgres pg_dump \
            -U "$PG_USER" \
            -d "$PG_DB" \
            --format=plain \
            --no-owner \
            --no-privileges \
            > "$BACKUP_FILE"
    fi

    local size=$(du -h "$BACKUP_FILE" | cut -f1)
    log_info "PostgreSQL 备份完成: $BACKUP_FILE ($size)"
}

# ---------- 备份 MySQL ----------
backup_mysql() {
    local dsn="${SQL_DSN:-}"
    if [[ -z "$dsn" ]]; then
        log_error "SQL_DSN 环境变量未设置，无法备份 MySQL"
        exit 1
    fi
    parse_mysql_dsn "$dsn"
    BACKUP_FILE="${BACKUP_FILE}.sql"
    log_info "正在备份 MySQL 数据库: ${MYSQL_USER}@${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB}"

    if command -v mysqldump &>/dev/null; then
        mysqldump \
            -h "$MYSQL_HOST" \
            -P "$MYSQL_PORT" \
            -u "$MYSQL_USER" \
            -p"$MYSQL_PASS" \
            --single-transaction \
            --routines \
            --triggers \
            "$MYSQL_DB" > "$BACKUP_FILE"
    else
        log_info "未找到 mysqldump，尝试通过 Docker 容器执行..."
        docker exec mysql mysqldump \
            -u "$MYSQL_USER" \
            -p"$MYSQL_PASS" \
            --single-transaction \
            --routines \
            --triggers \
            "$MYSQL_DB" > "$BACKUP_FILE"
    fi

    local size=$(du -h "$BACKUP_FILE" | cut -f1)
    log_info "MySQL 备份完成: $BACKUP_FILE ($size)"
}

# ---------- 清理过期备份 ----------
cleanup_old_backups() {
    log_info "清理超过 ${BACKUP_RETAIN_DAYS} 天的备份文件..."
    local deleted=0
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS find 语法
        while IFS= read -r -d '' file; do
            rm -f "$file"
            log_info "已删除过期备份: $(basename "$file")"
            ((deleted++))
        done < <(find "$BACKUP_DIR" -name "backup_*" -type f -mtime +${BACKUP_RETAIN_DAYS} -print0)
    else
        # Linux find 语法
        while IFS= read -r -d '' file; do
            rm -f "$file"
            log_info "已删除过期备份: $(basename "$file")"
            ((deleted++))
        done < <(find "$BACKUP_DIR" -name "backup_*" -type f -mtime +${BACKUP_RETAIN_DAYS} -print0)
    fi
    if [[ $deleted -eq 0 ]]; then
        log_info "无过期备份需要清理"
    else
        log_info "共清理 ${deleted} 个过期备份"
    fi
}

# ---------- 验证备份文件完整性 ----------
verify_backup() {
    if [[ ! -f "$BACKUP_FILE" ]]; then
        log_error "备份文件不存在: $BACKUP_FILE"
        exit 1
    fi
    local size=$(stat -c%s "$BACKUP_FILE" 2>/dev/null || stat -f%z "$BACKUP_FILE" 2>/dev/null)
    if [[ "$size" -lt 100 ]]; then
        log_error "备份文件过小 (${size} 字节)，可能备份失败: $BACKUP_FILE"
        exit 1
    fi
    # SQL 文件检查是否包含表定义
    if [[ "$BACKUP_FILE" == *.sql ]]; then
        if ! grep -qiE "CREATE TABLE|COPY|INSERT" "$BACKUP_FILE" 2>/dev/null; then
            log_error "SQL 备份文件未包含表定义或数据，可能备份失败: $BACKUP_FILE"
            exit 1
        fi
    fi
    log_info "备份文件验证通过: $BACKUP_FILE (${size} 字节)"
}

# ---------- 主流程 ----------
main() {
    log_info "=========================================="
    log_info "数据库备份开始"
    log_info "项目目录: $PROJECT_DIR"
    log_info "备份目录: $BACKUP_DIR"
    log_info "保留天数: ${BACKUP_RETAIN_DAYS} 天"
    log_info "=========================================="

    local db_type
    db_type=$(detect_db_type)
    log_info "检测到数据库类型: $db_type"

    case "$db_type" in
        sqlite)
            backup_sqlite
            ;;
        postgres)
            backup_postgres
            ;;
        mysql)
            backup_mysql
            ;;
        *)
            log_error "无法识别数据库类型，请通过 DB_TYPE 环境变量指定: sqlite / postgres / mysql"
            exit 1
            ;;
    esac

    verify_backup
    cleanup_old_backups

    log_info "=========================================="
    log_info "备份流程全部完成"
    log_info "=========================================="
}

main "$@"
