# 数据库备份与恢复

本文档描述 Took/New-API 的数据库备份与恢复机制，覆盖三种支持的数据库类型（SQLite、PostgreSQL、MySQL）。

---

## 一、数据库类型与存储位置

| 数据库类型 | 默认存储位置 | 配置方式 |
|-----------|------------|---------|
| **SQLite**（默认） | `./one-api.db` | 不设 `SQL_DSN` 环境变量，或设置 `SQLITE_PATH` |
| **PostgreSQL** | Docker 卷 `pg_data` | `SQL_DSN=postgresql://user:pass@host:5432/db` |
| **MySQL** | Docker 卷 `mysql_data` | `SQL_DSN=root:pass@tcp(host:3306)/db` |

Docker Compose 默认使用 PostgreSQL，数据持久化在 `pg_data` 命名卷中。

---

## 二、备份方案

### 2.1 自动备份（推荐）

#### Docker 环境（db-backup 容器）

`docker-compose.yml` 已内置 `db-backup` 服务，每天凌晨 3:00 自动执行 PostgreSQL 备份：

```bash
# 启动所有服务（含自动备份）
docker compose up -d

# 备份文件存储位置
ls ./backups/
# backup_20260101_030000.sql
# backup_20260102_030000.sql
# ...
```

**自定义备份配置**（通过 `.env` 文件或环境变量）：

```env
# 备份保留天数（默认 7 天）
BACKUP_RETAIN_DAYS=7
```

**手动触发 Docker 备份**：

```bash
docker compose run --rm db-backup /scripts/docker-backup-cron.sh
```

#### 非 Docker 环境（crontab）

```bash
# 编辑 crontab
crontab -e

# 添加每天凌晨 3 点的备份任务
0 3 * * * cd /path/to/new-api && ./scripts/backup.sh >> ./logs/backup.log 2>&1
```

### 2.2 手动备份

```bash
# 自动检测数据库类型
./scripts/backup.sh

# 指定数据库类型
DB_TYPE=postgres ./scripts/backup.sh
DB_TYPE=mysql ./scripts/backup.sh
DB_TYPE=sqlite ./scripts/backup.sh

# 自定义备份目录和保留天数
BACKUP_DIR=/data/backups BACKUP_RETAIN_DAYS=30 ./scripts/backup.sh
```

### 2.3 备份文件命名与存储

- 命名格式：`backup_YYYYMMDD_HHMMSS.sql`（PostgreSQL/MySQL）或 `backup_YYYYMMDD_HHMMSS.db`（SQLite）
- 默认存储目录：`./backups/`
- 自动清理超过保留天期的备份文件（默认 7 天）

```
backups/
├── backup_20260101_030000.sql   # 7 天前（已自动删除）
├── backup_20260102_030000.sql
├── backup_20260103_030000.sql
├── backup_20260104_030000.sql
├── backup_20260105_030000.sql
├── backup_20260106_030000.sql
├── backup_20260107_030000.sql
└── backup_20260108_030000.sql   # 今天
```

### 2.4 备份完整性验证

备份脚本内置完整性检查：
- 文件大小校验（< 100 字节视为失败）
- SQL 文件内容校验（必须包含 `CREATE TABLE` / `COPY` / `INSERT`）
- 失败时退出码非零，便于 cron 监控

---

## 三、恢复方案

### 3.1 恢复前准备

1. **停止应用服务**（避免恢复期间有新写入）：
   ```bash
   docker compose stop new-api
   ```

2. **确认备份文件**：
   ```bash
   ls -lh ./backups/
   ```

### 3.2 执行恢复

```bash
# 恢复 PostgreSQL 备份
./scripts/restore.sh backups/backup_20260101_030000.sql

# 恢复 SQLite 备份
./scripts/restore.sh backups/backup_20260101_030000.db
```

脚本会：
1. 显示恢复确认提示（需输入 `yes` 确认）
2. **自动创建当前数据的快照**（防止误操作）
3. 执行恢复
4. 提示重启应用服务

### 3.3 恢复后操作

```bash
# 重启应用服务
docker compose start new-api
# 或
docker compose up -d
```

---

## 四、灾难恢复流程

### 场景：数据库完全损坏

1. **启动新的数据库容器**：
   ```bash
   docker compose up -d postgres redis
   ```

2. **恢复最近的有效备份**：
   ```bash
   ./scripts/restore.sh backups/backup_20260108_030000.sql
   ```

3. **启动应用**：
   ```bash
   docker compose up -d new-api
   ```

4. **验证数据完整性**：
   - 登录管理后台，检查用户/渠道/日志是否正常
   - 测试 API 调用确认功能正常

### 场景：误删数据

1. 恢复前会自动创建当前数据快照
2. 使用 `restore.sh` 恢复到误操作前的备份
3. 快照保留在 `./backups/pre_restore_*.sql`

---

## 五、备份文件安全建议

1. **异地备份**：定期将 `./backups/` 目录同步到异地存储（如 S3、OSS）
   ```bash
   # 示例：同步到 S3
   aws s3 sync ./backups/ s3://your-bucket/new-api-backups/
   ```

2. **加密敏感备份**：备份文件含用户数据，传输前应加密
   ```bash
   gpg --encrypt --recipient admin@example.com backup_20260108_030000.sql
   ```

3. **定期恢复演练**：每月在测试环境验证备份可恢复性

4. **监控备份状态**：检查 `./logs/backup.log` 确认备份成功执行

---

## 六、脚本说明

| 脚本 | 用途 | 运行环境 |
|------|------|---------|
| `scripts/backup.sh` | 通用备份脚本，自动检测数据库类型 | 宿主机 / Docker |
| `scripts/restore.sh` | 通用恢复脚本，需指定备份文件 | 宿主机 / Docker |
| `scripts/docker-backup-cron.sh` | Docker 容器内 cron 调用的备份脚本 | Docker 容器内 |

### 环境变量参考

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_TYPE` | 自动检测 | `sqlite` / `postgres` / `mysql` |
| `BACKUP_DIR` | `./backups` | 备份存储目录 |
| `BACKUP_RETAIN_DAYS` | `7` | 备份保留天数 |
| `SQLITE_PATH` | `./one-api.db` | SQLite 数据库文件路径 |
| `SQL_DSN` | - | PostgreSQL/MySQL 连接字符串 |
| `POSTGRES_PASSWORD` | `123456`（仅本地调试） | PostgreSQL 密码，生产环境必须通过 .env.prod 覆盖 |
| `MYSQL_PASSWORD` | `123456`（仅本地调试） | MySQL 密码，生产环境必须通过 .env.prod 覆盖 |

---

## 七、版本管理与镜像 Tag 策略

### 7.1 版本记录方式

| 记录维度 | 方法 | 说明 |
|---------|------|------|
| **代码版本** | Git tag（如 `v0.6.0`） | 每次发布打 Git tag，标记代码快照 |
| **镜像版本** | Docker image tag（如 `new-api:v0.6.0`） | 镜像 tag 与 Git tag 对应，禁止使用 `:latest` |
| **运行版本** | `common.Version` 编译时注入 | 通过 `go build -ldflags` 注入，可在 `/api/status` 查看 |
| **部署版本** | `.env.prod` 中 `NEW_API_VERSION` | docker-compose 据此 tag 拉取/构建镜像 |

### 7.2 镜像 Tag 命名规范

```
new-api:v<MAJOR>.<MINOR>.<PATCH>     # 正式版本（如 new-api:v0.6.0）
new-api:v<MAJOR>.<MINOR>.<PATCH>-rc<N>  # 发布候选（如 new-api:v0.6.0-rc1）
new-api:ci-check                      # CI 构建检查专用（不用于部署）
```

**禁止使用的 tag：**
- `:latest` — 无法追踪版本，无法回滚
- `:dev` / `:test` — 语义模糊，可能被意外覆盖

### 7.3 构建与推送流程

```bash
# 1. 更新 VERSION 文件
echo "v0.6.1" > VERSION

# 2. 构建镜像（使用 docker-compose 或手动）
docker compose --env-file .env.prod build new-api

# 3. 验证镜像
docker image inspect new-api:v0.6.1 --format '{{.Id}} {{.Config.User}}'

# 4. 打 Git tag
git tag v0.6.1
git push origin v0.6.1
```

---

## 八、部署流程

### 8.1 标准部署步骤

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 执行部署前检查（必须通过）
./scripts/pre-deploy-check.sh .env.prod

# 3. 备份数据库（部署前快照）
./scripts/backup.sh

# 4. 构建新版本镜像
docker compose --env-file .env.prod build new-api

# 5. 滚动重启（先停止旧容器再启动新容器）
docker compose --env-file .env.prod up -d new-api

# 6. 验证服务健康
curl -f http://localhost:3000/health || echo "健康检查失败！"
curl -f http://localhost:3000/api/status | grep '"success":true'

# 7. 确认版本号
curl http://localhost:3000/api/status | grep version
```

### 8.2 部署前检查脚本

`scripts/pre-deploy-check.sh` 在部署前自动验证：

1. **环境变量完整性** — 必需变量已设置、密码不为弱默认值、SESSION_SECRET 长度足够
2. **数据库连通性** — PostgreSQL 容器可连接
3. **Redis 连通性** — Redis 容器可连接且密码正确
4. **Docker 可用性** — docker daemon 运行正常、docker compose 可用
5. **磁盘空间** — 至少 2GB 可用空间

**任何 FAIL 项都会阻止部署**（退出码 1）。

```bash
# 用法
./scripts/pre-deploy-check.sh .env.prod
```

---

## 九、回滚方案

### 9.1 回滚决策流程

```
部署后发现异常
    │
    ├── 能否快速修复？（如配置错误）
    │       ├── 是 → 修复配置，重启服务
    │       └── 否 → 继续向下
    │
    ├── 是否有数据破坏？
    │       ├── 是 → 先恢复数据库（见 9.4），再回滚镜像
    │       └── 否 → 直接回滚镜像（见 9.3）
    │
    └── 回滚后验证 → 确认服务恢复正常
```

### 9.2 快速恢复旧版本（镜像回滚）

```bash
# 1. 确认当前运行版本
docker inspect new-api --format '{{.Config.Image}}'
# 输出示例: new-api:v0.6.1

# 2. 确认要回滚到的旧版本镜像仍存在
docker images new-api --format '{{.Tag}}  {{.ID}}  {{.CreatedAt}}'
# v0.6.1   a1b2c3d4...   2026-07-15 10:00
# v0.6.0   e5f6g7h8...   2026-07-10 14:00  ← 回滚目标

# 3. 修改 .env.prod 中的版本号
sed -i 's/NEW_API_VERSION=v0.6.1/NEW_API_VERSION=v0.6.0/' .env.prod

# 4. 使用旧版本镜像重启（不重新构建）
docker compose --env-file .env.prod up -d new-api

# 5. 验证回滚成功
curl http://localhost:3000/api/status | grep version
docker inspect new-api --format '{{.Config.Image}}'
# 应显示 new-api:v0.6.0
```

**如果旧版本镜像已被删除：**

```bash
# 需要从 Git tag 重新构建
git checkout v0.6.0
NEW_API_VERSION=v0.6.0 docker compose --env-file .env.prod build new-api
docker compose --env-file .env.prod up -d new-api
git checkout main  # 切回主分支
```

### 9.3 数据库迁移回滚注意事项

> ⚠️ **数据库回滚是高风险操作，必须先备份！**

Took/New-API 使用 GORM `AutoMigrate` 自动迁移，迁移特点与限制：

| 特性 | 说明 |
|------|------|
| 迁移方式 | `AutoMigrate` — 只增不删（添加新列/新表，不删除旧列/旧表） |
| 回滚兼容性 | **向前兼容** — 新版本添加的列在旧版本中被忽略，不影响旧版本运行 |
| 风险等级 | 低 — 无破坏性 schema 变更时可直接回滚镜像 |

**回滚前的数据库检查清单：**

1. **确认无破坏性迁移**：
   ```bash
   # 查看新版本是否有手动迁移脚本或 RenameColumn/DropColumn 调用
   git log v0.6.0..v0.6.1 --oneline -- model/
   git diff v0.6.0..v0.6.1 -- model/ | grep -E "DropColumn|RenameColumn|DropTable"
   ```

2. **如果存在破坏性迁移（删列/改类型）**：
   - **禁止直接回滚镜像** — 旧代码可能依赖已删除的列
   - 需要先从备份恢复数据库到迁移前状态：
     ```bash
     docker compose stop new-api
     ./scripts/restore.sh backups/backup_YYYYMMDD_HHMMSS.sql
     ```
   - 再回滚镜像

3. **如果仅添加了新列/新表（向前兼容）**：
   - 可直接回滚镜像，数据库无需恢复
   - 旧版本会忽略新增的列，不影响运行

### 9.4 数据库回滚步骤

```bash
# 1. 停止应用服务
docker compose --env-file .env.prod stop new-api

# 2. 确认备份文件（选择部署前的备份）
ls -lh ./backups/

# 3. 执行恢复（脚本会自动创建当前数据快照）
./scripts/restore.sh backups/backup_YYYYMMDD_HHMMSS.sql

# 4. 回滚镜像版本
sed -i 's/NEW_API_VERSION=v0.6.1/NEW_API_VERSION=v0.6.0/' .env.prod
docker compose --env-file .env.prod up -d new-api

# 5. 验证数据完整性
curl http://localhost:3000/health
curl http://localhost:3000/api/status | grep '"success":true'
```

### 9.5 回滚后操作

1. **通知相关方**：记录回滚原因、时间、回滚到的版本
2. **保留故障版本镜像**：不要立即删除，便于后续问题排查
3. **分析根因**：在测试环境复现问题，修复后再重新发布
4. **更新部署日志**：记录此次回滚事件

---

## 十、环境配置隔离

项目通过 `.env.*` 文件实现环境隔离，模板文件提交到仓库，真实配置文件被 gitignore 排除。

| 文件 | 用途 | 是否提交仓库 |
|------|------|-------------|
| `.env.example` | 通用配置参考 | ✅ 是 |
| `.env.dev.example` | 开发环境模板 | ✅ 是 |
| `.env.test.example` | 测试环境模板 | ✅ 是 |
| `.env.prod.example` | 生产环境模板 | ✅ 是 |
| `.env.dev` | 开发环境真实配置 | ❌ 否（gitignored） |
| `.env.test` | 测试环境真实配置 | ❌ 否（gitignored） |
| `.env.prod` | 生产环境真实配置 | ❌ 否（gitignored） |

### 加载方式

```bash
# 开发环境
docker compose --env-file .env.dev up -d

# 测试环境
docker compose --env-file .env.test up -d

# 生产环境
docker compose --env-file .env.prod up -d
```

### 初始化

```bash
# 从模板创建真实配置文件
cp .env.prod.example .env.prod
# 编辑 .env.prod，填入真实强密码与密钥
vim .env.prod
```
