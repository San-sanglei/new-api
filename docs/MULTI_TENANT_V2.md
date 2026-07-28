# SaaS 多租户隔离方案 V2（第一阶段）

## 设计目标

为 new-api 增加 SaaS 多租户数据隔离能力，采用渐进式迁移，保留所有现有功能。

## 设计原则

1. **保留现有 Role 体系**：RoleGuest=0 / RoleUser=1 / RoleAdmin=10 / RoleRoot=100，不新增 `is_tenant_admin` 字段
2. **多对多关系**：使用 `tenant_members` 表维护用户与租户的关系，一个用户可加入多个租户
3. **当前活动租户**：`users.current_tenant_id` 表示用户当前活动租户，支持后续切换
4. **不破坏现有功能**：所有 API 返回格式不变，不修改支付逻辑，不修改 Epay
5. **渐进式迁移**：第一阶段只做基础结构，第二阶段起才隔离业务表

## 第一阶段范围

### 新增文件

| 文件 | 用途 |
|---|---|
| `model/tenant.go` | Tenant 模型 + 默认租户初始化 |
| `model/tenant_member.go` | TenantMember 模型 + 成员管理 + GetUserCurrentTenantId |
| `model/migration_tenant.go` | 幂等迁移逻辑（跨 SQLite/MySQL/PostgreSQL） |
| `service/tenant_service.go` | GetTenantID / IsSuperAdmin / CanAccessUserData |
| `docs/MULTI_TENANT_V2.md` | 本文档 |

### 修改文件

| 文件 | 改动 |
|---|---|
| `model/user.go` | User 结构体加 `CurrentTenantID int` 字段（column: current_tenant_id, default: 1） |
| `model/main.go` | `migrateDB()` 末尾调用 `migrateTenantSchema()` |
| `constant/context_key.go` | 新增 `ContextKeyTenantId = "tenant_id"` |
| `middleware/auth.go` | `authHelper` 末尾写入 tenant_id 到 context |

### 不修改的内容

- ❌ 不修改任何 controller
- ❌ 不修改 router/main.go
- ❌ 不修改支付逻辑（Epay/Stripe/Creem/Waffo）
- ❌ 不修改业务表（tokens/logs/top_ups/subscriptions/channels/models/vendors/pricing）
- ❌ 不注册 TenantContext 全局中间件（只在 auth 成功后写入 context）
- ❌ 不创建 migrations SQL 目录
- ❌ 不提交 git commit

## 数据库模型

### tenants 表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | int PK | 主键 |
| name | varchar(128) | 租户名称（不唯一，允许同名） |
| slug | varchar(64) | URL 友好标识（唯一） |
| status | int | 1=活跃 / 2=暂停 / 3=删除 |
| owner_user_id | int | 所有者用户 ID |
| created_at | bigint | 创建时间 |
| updated_at | bigint | 更新时间 |
| deleted_at | bigint | 软删除 |

**默认租户**：id=1, name="Took Official", slug="took-official", owner=1（root）

### tenant_members 表

| 字段 | 类型 | 说明 |
|---|---|---|
| id | int PK | 主键 |
| tenant_id | int | 租户 ID |
| user_id | int | 用户 ID |
| role | int | 1=成员 / 2=管理员 / 3=所有者 |
| status | int | 1=活跃 / 2=已邀请 / 3=已禁用 / 4=已退出 |
| joined_at | bigint | 加入时间 |
| created_at | bigint | 创建时间 |
| updated_at | bigint | 更新时间 |

**唯一约束**：(tenant_id, user_id)

### users 表变更

新增列：`current_tenant_id INT DEFAULT 1`

**迁移逻辑**：所有历史用户 `current_tenant_id=1`（默认租户），并自动加入 `tenant_members` 表作为默认租户的成员。

## Context 机制

### 写入路径

```
请求进入
  ↓
middleware/auth.go 的 authHelper
  ↓
认证成功（session 或 access token）
  ↓
获取 user_id
  ↓
调用 model.GetUserCurrentTenantId(uid)
  ↓
写入 c.Set("tenant_id", tenantId)
  ↓
后续 handler 通过 c.GetInt("tenant_id") 或 service.GetTenantID(c) 读取
```

### 读取方式

```go
// 推荐方式
tenantId := service.GetTenantID(c)

// 或直接读取
tenantId := c.GetInt(string(constant.ContextKeyTenantId))
```

### 回退策略

- 未认证请求：不写入 tenant_id（默认 0）
- 用户不存在：回退到 `DefaultTenantID=1`
- 用户不在任何租户中：回退到 `DefaultTenantID=1`
- 查询失败：回退到 `DefaultTenantID=1`，不阻塞请求

## 迁移流程

### 自动迁移（启动时执行）

`model/main.go` 的 `migrateDB()` 末尾调用 `migrateTenantSchema()`：

1. `DB.AutoMigrate(&Tenant{}, &TenantMember{})` 创建表
2. `EnsureDefaultTenant()` 创建默认租户（id=1）
3. `addColumnIfNotExists("users", "current_tenant_id", "INT DEFAULT 1")` 幂等加列
4. `migrateExistingUsersToDefaultTenant()` 将现有用户加入默认租户

### 幂等性

- 所有迁移操作均幂等，可重复执行
- `addColumnIfNotExists` 按数据库类型检查列是否存在
- `migrateExistingUsersToDefaultTenant` 用 LEFT JOIN 找出未迁移的用户

## 权限模型

### 系统级 Role（保留不变）

| Role | 值 | 说明 |
|---|---|---|
| RoleGuestUser | 0 | 访客 |
| RoleCommonUser | 1 | 普通用户 |
| RoleAdminUser | 10 | 管理员 |
| RoleRootUser | 100 | 超级管理员（可跨租户） |

### 租户内 Role（tenant_members.role）

| Role | 值 | 说明 |
|---|---|---|
| TenantMemberRoleMember | 1 | 普通成员 |
| TenantMemberRoleAdmin | 2 | 租户管理员 |
| TenantMemberRoleOwner | 3 | 租户所有者 |

### 权限矩阵

| 场景 | 系统级 Role | 租户内 Role | 可见范围 |
|---|---|---|---|
| 普通用户 | RoleUser=1 | member | 自己的数据 |
| 租户管理员 | RoleUser=1 | admin | 本租户所有用户数据 |
| 普通管理员 | RoleAdmin=10 | - | 本租户数据（保持现有行为） |
| 超级管理员 | RoleRoot=100 | - | 所有租户（跨租户） |

**注意**：第一阶段不实现权限校验逻辑，`CanAccessUserData` 始终返回 true，保持向后兼容。第二阶段才追加 tenant_id 校验。

## 回滚方案

### SQL 回滚

```sql
-- 1. 删除 users 表新增列
ALTER TABLE users DROP COLUMN IF EXISTS current_tenant_id;

-- 2. 删除 tenant_members 表
DROP TABLE IF EXISTS tenant_members;

-- 3. 删除 tenants 表
DROP TABLE IF EXISTS tenants;
```

### 代码回滚

```bash
git revert <tenant-commit-hash>
```

### 应用回滚步骤

1. 部署回滚代码版本
2. 执行 SQL 回滚脚本
3. 重启服务
4. 验证登录、充值、API 调用正常

## 风险分析

| 风险 | 等级 | 缓解措施 |
|---|---|---|
| Migration 失败 | 低 | 所有操作幂等，可重复执行 |
| users 表锁 | 低 | `ALTER TABLE ADD COLUMN ... DEFAULT 1` 在 PG 中秒级完成 |
| 现有 API 返回变化 | 低 | User JSON 多了 `current_tenant_id` 字段，前端不依赖 |
| 中间件性能 | 低 | 每次请求多一次 DB 查询，可后续加 Redis 缓存 |
| 支付逻辑误改 | 无 | 严格限制修改文件列表，不碰支付代码 |
| Epay 回调异常 | 无 | Epay 代码完全不动 |

## 后续阶段

### 第二阶段：数据隔离

- tokens / logs / top_ups / subscriptions 表加 tenant_id
- Model 层查询追加 `WHERE tenant_id = ?`
- Controller 层追加跨租户访问校验
- 实现 `CanAccessUserData` 完整逻辑

### 第三阶段：配置隔离

- channels / models / vendors / pricing 按租户配置
- 租户管理 API（CRUD 租户、成员）
- 租户切换功能
- 租户级配额/计费
