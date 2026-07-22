#!/usr/bin/env bash
# =============================================================================
# 部署前检查脚本 (Pre-Deploy Check)
# =============================================================================
# 用途：在执行生产部署前，验证所有关键依赖与配置是否就绪。
#       任何一项检查失败即阻止部署（退出码 1）。
#
# 用法：
#   ./scripts/pre-deploy-check.sh                    # 使用默认 .env
#   ./scripts/pre-deploy-check.sh .env.prod          # 指定环境文件
#   ENV_FILE=.env.prod ./scripts/pre-deploy-check.sh # 通过环境变量指定
#
# 检查项：
#   1. 环境变量是否完整（必需变量已设置、密码不为弱默认值）
#   2. 数据库是否可连接
#   3. Redis 是否可连接
#   4. Docker 与 Docker Compose 是否可用
#   5. 磁盘空间是否足够（至少 2GB 可用）
# =============================================================================

set -euo pipefail

# ---- 颜色输出 ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0
WARN=0

print_pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASS=$((PASS + 1)); }
print_fail() { echo -e "${RED}[FAIL]${NC} $1"; FAIL=$((FAIL + 1)); }
print_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; WARN=$((WARN + 1)); }

# ---- 加载环境文件 ----
ENV_FILE="${1:-${ENV_FILE:-.env}}"
if [[ -f "$ENV_FILE" ]]; then
    echo "==> 加载环境文件: $ENV_FILE"
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
else
    echo "==> 未找到环境文件 $ENV_FILE，使用当前环境变量"
fi

echo "========================================"
echo "  Took/New-API 部署前检查"
echo "========================================"
echo ""

# =============================================================================
# 检查 1：环境变量完整性
# =============================================================================
echo "==> [1/5] 检查环境变量..."

REQUIRED_VARS=(
    "SQL_DSN"
    "REDIS_CONN_STRING"
    "SESSION_SECRET"
)

OPTIONAL_VARS=(
    "POSTGRES_PASSWORD"
    "REDIS_PASSWORD"
    "NODE_NAME"
    "TZ"
)

for var in "${REQUIRED_VARS[@]}"; do
    val="${!var:-}"
    if [[ -z "$val" ]]; then
        print_fail "必需环境变量 $var 未设置"
    else
        print_pass "必需环境变量 $var 已设置"
    fi
done

for var in "${OPTIONAL_VARS[@]}"; do
    val="${!var:-}"
    if [[ -z "$val" ]]; then
        print_warn "可选环境变量 $var 未设置（生产环境建议设置）"
    else
        print_pass "可选环境变量 $var 已设置"
    fi
done

# 检查密码是否为弱默认值
if [[ "${POSTGRES_PASSWORD:-}" == "123456" ]]; then
    print_fail "POSTGRES_PASSWORD 仍为弱默认值 123456，生产环境必须修改"
fi
if [[ "${REDIS_PASSWORD:-}" == "123456" ]]; then
    print_fail "REDIS_PASSWORD 仍为弱默认值 123456，生产环境必须修改"
fi
if [[ "${MYSQL_PASSWORD:-}" == "123456" ]]; then
    print_fail "MYSQL_PASSWORD 仍为弱默认值 123456，生产环境必须修改"
fi

# 检查 SESSION_SECRET 是否为默认占位符
if [[ "${SESSION_SECRET:-}" == "CHANGE_ME_TO_RANDOM_64_CHAR_HEX_STRING" ]] || \
   [[ "${SESSION_SECRET:-}" == "random_string" ]] || \
   [[ "${SESSION_SECRET:-}" == "dev-session-secret-do-not-use-in-prod" ]]; then
    print_fail "SESSION_SECRET 仍为模板占位符，生产环境必须替换为随机字符串"
fi

# 检查 SESSION_SECRET 长度（至少 32 字符）
if [[ -n "${SESSION_SECRET:-}" ]] && [[ ${#SESSION_SECRET} -lt 32 ]]; then
    print_fail "SESSION_SECRET 长度不足 32 字符（当前 ${#SESSION_SECRET} 字符），建议使用 openssl rand -hex 32 生成"
fi

# 生产环境 COOKIE_SECURE 检查
if [[ "${COOKIE_SECURE:-}" != "true" ]]; then
    print_warn "COOKIE_SECURE 未设为 true，生产环境（HTTPS）应启用"
fi

echo ""

# =============================================================================
# 检查 2：数据库连通性
# =============================================================================
echo "==> [2/5] 检查数据库连通性..."

if [[ -z "${SQL_DSN:-}" ]]; then
    print_fail "SQL_DSN 未设置，无法检查数据库连通性"
else
    # 尝试使用 docker exec 检查 PostgreSQL（如果 docker 可用且容器在运行）
    if command -v docker &>/dev/null; then
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^postgres"; then
            if docker exec postgres pg_isready -U root -d new-api 2>/dev/null; then
                print_pass "PostgreSQL 容器 (postgres) 可连接"
            else
                print_fail "PostgreSQL 容器 (postgres) 不可连接"
            fi
        else
            print_warn "PostgreSQL 容器 (postgres) 未运行，跳过容器内连通性检查"
        fi
    else
        print_warn "docker 命令不可用，跳过数据库容器连通性检查"
    fi
fi

echo ""

# =============================================================================
# 检查 3：Redis 连通性
# =============================================================================
echo "==> [3/5] 检查 Redis 连通性..."

if [[ -z "${REDIS_CONN_STRING:-}" ]]; then
    print_fail "REDIS_CONN_STRING 未设置，无法检查 Redis 连通性"
else
    if command -v docker &>/dev/null; then
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^redis"; then
            REDIS_PASS="${REDIS_PASSWORD:-}"
            if [[ -n "$REDIS_PASS" ]]; then
                if docker exec redis redis-cli -a "$REDIS_PASS" ping 2>/dev/null | grep -q "PONG"; then
                    print_pass "Redis 容器 (redis) 可连接"
                else
                    print_fail "Redis 容器 (redis) 不可连接或密码错误"
                fi
            else
                if docker exec redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
                    print_pass "Redis 容器 (redis) 可连接"
                else
                    print_fail "Redis 容器 (redis) 不可连接"
                fi
            fi
        else
            print_warn "Redis 容器 (redis) 未运行，跳过容器内连通性检查"
        fi
    else
        print_warn "docker 命令不可用，跳过 Redis 容器连通性检查"
    fi
fi

echo ""

# =============================================================================
# 检查 4：Docker 与 Docker Compose 可用性
# =============================================================================
echo "==> [4/5] 检查 Docker 可用性..."

if ! command -v docker &>/dev/null; then
    print_fail "docker 命令未安装或不在 PATH 中"
else
    print_pass "docker 命令可用"
fi

if ! docker info &>/dev/null 2>&1; then
    print_fail "docker daemon 未运行或当前用户无权限访问（请检查 docker 组或 sudo）"
else
    print_pass "docker daemon 运行正常"
fi

# 检查 docker compose（v2）或 docker-compose（v1）
if docker compose version &>/dev/null 2>&1; then
    print_pass "docker compose (v2) 可用"
elif command -v docker-compose &>/dev/null; then
    print_pass "docker-compose (v1) 可用"
else
    print_fail "docker compose / docker-compose 均不可用"
fi

# 检查 docker-compose.yml 是否存在
if [[ -f "docker-compose.yml" ]]; then
    print_pass "docker-compose.yml 存在"
else
    print_fail "docker-compose.yml 不存在"
fi

echo ""

# =============================================================================
# 检查 5：磁盘空间
# =============================================================================
echo "==> [5/5] 检查磁盘空间..."

# 最低要求：2GB 可用空间
MIN_DISK_MB=2048

if [[ "$(uname)" == "Darwin" ]] || [[ "$(uname)" == "Linux" ]]; then
    # 获取当前目录所在分区的可用空间（MB）
    AVAILABLE_MB=$(df -m . | awk 'NR==2 {print $4}')

    if [[ "$AVAILABLE_MB" -ge "$MIN_DISK_MB" ]]; then
        print_pass "磁盘空间充足：${AVAILABLE_MB}MB 可用（要求 ≥ ${MIN_DISK_MB}MB）"
    else
        print_fail "磁盘空间不足：${AVAILABLE_MB}MB 可用（要求 ≥ ${MIN_DISK_MB}MB）"
    fi

    # Docker 磁盘使用情况
    if docker info &>/dev/null 2>&1; then
        DOCKER_ROOT=$(docker info --format '{{.DockerRootDir}}' 2>/dev/null || echo "/var/lib/docker")
        if [[ -d "$DOCKER_ROOT" ]]; then
            DOCKER_AVAIL_MB=$(df -m "$DOCKER_ROOT" | awk 'NR==2 {print $4}')
            if [[ "$DOCKER_AVAIL_MB" -ge "$MIN_DISK_MB" ]]; then
                print_pass "Docker 目录磁盘空间充足：${DOCKER_AVAIL_MB}MB 可用"
            else
                print_fail "Docker 目录磁盘空间不足：${DOCKER_AVAIL_MB}MB 可用（要求 ≥ ${MIN_DISK_MB}MB）"
            fi
        fi
    fi
else
    print_warn "非 Linux/macOS 系统，跳过磁盘空间检查"
fi

echo ""

# =============================================================================
# 汇总
# =============================================================================
echo "========================================"
echo "  检查结果汇总"
echo "========================================"
echo -e "  ${GREEN}通过: $PASS${NC}"
echo -e "  ${YELLOW}警告: $WARN${NC}"
echo -e "  ${RED}失败: $FAIL${NC}"
echo "========================================"

if [[ "$FAIL" -gt 0 ]]; then
    echo ""
    echo -e "${RED}部署前检查未通过！${NC} 请修复上述 FAIL 项后再执行部署。"
    exit 1
fi

if [[ "$WARN" -gt 0 ]]; then
    echo ""
    echo -e "${YELLOW}存在警告项，请确认无影响后继续部署。${NC}"
fi

echo ""
echo -e "${GREEN}部署前检查通过！可以执行部署。${NC}"
exit 0
