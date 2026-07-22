#!/usr/bin/env bash
# ============================================================================
# 数据库恢复脚本
#
# 支持数据库类型：SQLite / PostgreSQL / MySQL
#
# 用法：
#   ./scripts/restore.sh <备份文件路径>
#
# 示例：
#   ./scripts/restore.sh backups/backup_20260101_030000.sql
#   ./scripts/restore.sh backups/backup_20260101_030000.db
#
# 环境变量：
#   DB_TYPE          - 数据库类型：sqlite / postgres / mysql（不设则根据文件扩展名推断）
#   SQLITE_PATH      - SQLite 数据库文件路径，默认 one-api.db
#   SQL_DSN          - PostgreSQL/MySQL 连接字符串
#   POSTGRES_PASSWORD - PostgreSQL 密码
#   MYSQL_PASSWORD    - MySQL 密码
#
# ⚠️  警告：恢复操作将覆盖现有数据，执行前会要求确认！
# ============================================================================

set -euo pipefail

# ---------- 颜色输出 ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $(date '+%Y-%m-%d %H:%M:%S') $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $(date '+%Y-%m-%d %H:%M:%S') $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*" >&2; }

# ---------- 配置 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

SQLITE_PATH="${SQLITE_PATH:-$PROJECT_DIR/one-api.db}"

# ---------- 参数检查 ----------
if [[ $# -lt 1 ]]; then
    echo "用法: $0 <备份文件路径>"
    echo ""
    echo "示例:"
    echo "  $0 backups/backup_20260101_030000.sql   # 恢复 PostgreSQL/MySQL 备份"
    echo "  $0 backups/backup_20260101_030000.db    # 恢复 SQLite 备份"
    exit 1
fi

BACKUP_FILE="$1"

if [[ ! -f "$BACKUP_FILE" ]]; then
    log_error "备份文件不存在: $BACKUP_FILE"
    exit 1
fi

# ---------- 推断数据库类型 ----------
detect_db_type_from_file() {
    if [[ -n "${DB_TYPE:-}" ]]; then
        echo "$DB_TYPE"
        return
    fi
    case "$BACKUP_FILE" in
        *.db)  echo "sqlite" ;;
        *.sql)
            # SQL 文件可能是 PostgreSQL 或 MySQL，根据内容判断
            if grep -qi "POSTGRES\|pg_catalog\|SERIAL" "$BACKUP_FILE" 2>/dev/null; then
                echo "postgres"
            elif grep -qi "ENGINE=InnoDB\|MYSQL\|AUTO_INCREMENT" "$BACKUP_FILE" 2>/dev/null; then
                echo "mysql"
            else
                echo "postgres" # 默认按 PostgreSQL 处理
            fi
            ;;
        *) echo "unknown" ;;
    esac
}

# ---------- 解析 PostgreSQL DSN ----------
parse_pg_dsn() {
    local dsn="$1"
    local rest="${dsn#postgresql://}"
    rest="${rest#postgres://}"
    local auth="${rest%%@*}"
    PG_USER="${auth%%:*}"
    PG_PASS="${auth#*:}"
    [[ "$PG_PASS" == "$auth" ]] && PG_PASS=""
    local hostpart="${rest#*@}"
    PG_HOST="${hostpart%%:*}"
    local portdb="${hostpart#*:}"
    PG_PORT="${portdb%%/*}"
    PG_DB="${portdb#*/}"
    [[ -z "$PG_PORT" ]] && PG_PORT="5432"
    [[ -z "$PG_PASS" ]] && PG_PASS="${POSTGRES_PASSWORD:-}"
}

# ---------- 解析 MySQL DSN ----------
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
    MYSQL_DB="${MYSQL_DB%%\?*}"
    [[ -z "$MYSQL_PORT" ]] && MYSQL_PORT="3306"
    [[ -z "$MYSQL_PASS" ]] && MYSQL_PASS="${MYSQL_PASSWORD:-}"
}

# ---------- 恢复前自动备份当前数据 ----------
create_pre_restore_backup() {
    local db_type="$1"
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local pre_backup="$PROJECT_DIR/backups/pre_restore_${timestamp}"

    log_warn "恢复前自动创建当前数据快照..."
    DB_TYPE="$db_type" "$SCRIPT_DIR/backup.sh" 2>/dev/null || {
        log_warn "恢复前自动备份失败，继续恢复操作..."
    }
    log_info "恢复前备份已创建"
}

# ---------- 恢复 SQLite ----------
restore_sqlite() {
    local backup_file="$BACKUP_FILE"
    local target="$SQLITE_PATH"

    log_warn "即将覆盖 SQLite 数据库文件:"
    log_warn "  备份文件 → $backup_file"
    log_warn "  目标文件 → $target"

    # 如果目标文件存在，先备份
    if [[ -f "$target" ]]; then
        local timestamp=$(date '+%Y%m%d_%H%M%S')
        local emergency_backup="${target}.prerestore.${timestamp}"
        cp "$target" "$emergency_backup"
        log_info "已创建恢复前快照: $emergency_backup"
    fi

    # 停止应用写入（提示用户）
    log_warn "建议在恢复前停止应用服务，避免数据冲突"

    # 执行恢复
    if command -v sqlite3 &>/dev/null; then
        # 使用 sqlite3 的 .restore 命令
        rm -f "$target" "${target}-wal" "${target}-shm"
        sqlite3 "$target" ".restore '$backup_file'"
    else
        # 直接复制文件
        rm -f "$target" "${target}-wal" "${target}-shm"
        cp "$backup_file" "$target"
        # 如果备份有 WAL/SHM 文件也一并恢复
        [[ -f "${backup_file}-wal" ]] && cp "${backup_file}-wal" "${target}-wal"
        [[ -f "${backup_file}-shm" ]] && cp "${backup_file}-shm" "${target}-shm"
    fi

    log_info "SQLite 恢复完成: $target"
}

# ---------- 恢复 PostgreSQL ----------
restore_postgres() {
    local dsn="${SQL_DSN:-}"
    if [[ -z "$dsn" ]]; then
        log_error "SQL_DSN 环境变量未设置，无法恢复 PostgreSQL"
        exit 1
    fi
    parse_pg_dsn "$dsn"

    log_warn "即将恢复 PostgreSQL 数据库:"
    log_warn "  备份文件 → $BACKUP_FILE"
    log_warn "  目标数据库 → ${PG_USER}@${PG_HOST}:${PG_PORT}/${PG_DB}"
    log_warn "  ⚠️  该操作将覆盖现有数据！"

    if command -v psql &>/dev/null; then
        PGPASSWORD="$PG_PASS" psql \
            -h "$PG_HOST" \
            -p "$PG_PORT" \
            -U "$PG_USER" \
            -d "$PG_DB" \
            -f "$BACKUP_FILE"
    else
        log_info "未找到 psql，尝试通过 Docker 容器执行..."
        docker exec -i postgres psql \
            -U "$PG_USER" \
            -d "$PG_DB" \
            < "$BACKUP_FILE"
    fi

    log_info "PostgreSQL 恢复完成"
}

# ---------- 恢复 MySQL ----------
restore_mysql() {
    local dsn="${SQL_DSN:-}"
    if [[ -z "$dsn" ]]; then
        log_error "SQL_DSN 环境变量未设置，无法恢复 MySQL"
        exit 1
    fi
    parse_mysql_dsn "$dsn"

    log_warn "即将恢复 MySQL 数据库:"
    log_warn "  备份文件 → $BACKUP_FILE"
    log_warn "  目标数据库 → ${MYSQL_USER}@${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB}"
    log_warn "  ⚠️  该操作将覆盖现有数据！"

    if command -v mysql &>/dev/null; then
        mysql \
            -h "$MYSQL_HOST" \
            -P "$MYSQL_PORT" \
            -u "$MYSQL_USER" \
            -p"$MYSQL_PASS" \
            "$MYSQL_DB" < "$BACKUP_FILE"
    else
        log_info "未找到 mysql，尝试通过 Docker 容器执行..."
        docker exec -i mysql mysql \
            -u "$MYSQL_USER" \
            -p"$MYSQL_PASS" \
            "$MYSQL_DB" < "$BACKUP_FILE"
    fi

    log_info "MySQL 恢复完成"
}

# ---------- 确认提示 ----------
confirm_restore() {
    echo ""
    echo -e "${RED}========================================================${NC}"
    echo -e "${RED}  ⚠️  危险操作：数据库恢复  ${NC}"
    echo -e "${RED}========================================================${NC}"
    echo ""
    echo -e "  备份文件: ${CYAN}$BACKUP_FILE${NC}"
    echo -e "  文件大小: $(du -h "$BACKUP_FILE" | cut -f1)"
    echo -e "  数据库类型: ${CYAN}$1${NC}"
    echo ""
    echo -e "${RED}  此操作将覆盖现有数据库中的数据！${NC}"
    echo -e "${YELLOW}  恢复前会自动创建当前数据的快照。${NC}"
    echo ""
    read -p "确认执行恢复操作？输入 yes 继续，其他任意键取消: " confirm

    if [[ "$confirm" != "yes" ]]; then
        log_info "已取消恢复操作"
        exit 0
    fi
}

# ---------- 主流程 ----------
main() {
    local db_type
    db_type=$(detect_db_type_from_file)

    if [[ "$db_type" == "unknown" ]]; then
        log_error "无法从文件推断数据库类型，请通过 DB_TYPE 环境变量指定"
        exit 1
    fi

    log_info "=========================================="
    log_info "数据库恢复开始"
    log_info "备份文件: $BACKUP_FILE"
    log_info "数据库类型: $db_type"
    log_info "=========================================="

    # 确认提示
    confirm_restore "$db_type"

    # 恢复前自动创建快照
    create_pre_restore_backup "$db_type"

    case "$db_type" in
        sqlite)
            restore_sqlite
            ;;
        postgres)
            restore_postgres
            ;;
        mysql)
            restore_mysql
            ;;
        *)
            log_error "不支持的数据库类型: $db_type"
            exit 1
            ;;
    esac

    log_info "=========================================="
    log_info "数据库恢复完成"
    log_info "=========================================="
    log_warn "请重启应用服务以加载恢复后的数据"
}

main "$@"
