# Production Release Checklist

## 部署前

### 安全配置

- [ ] 修改数据库密码（POSTGRES_PASSWORD 设置为强密码，docker-compose 未设置时会 fail-closed）
- [ ] 修改 Redis 密码（REDIS_PASSWORD 设置为强密码，docker-compose 未设置时会 fail-closed）
- [ ] 配置 SESSION_SECRET（多节点部署必填，否则 session 无法跨节点共享）
- [ ] 配置 webhook secret（StripeWebhookSecret / CreemWebhookSecret / Waffo webhook secret）

### 数据持久化

- [ ] 数据库备份验证（docker-compose db-backup 每日 03:00，验证备份文件可恢复）
- [ ] Redis 持久化检查（确认 AOF/RDB 策略，避免重启丢失缓存）
- [ ] 数据卷挂载验证（./data + ./logs 挂载到持久化存储）

### 版本管理

- [ ] git tag 发布版本（项目必须纳入 git 版本控制，部署前打 tag）
- [ ] 记录当前 commit hash 便于回滚
- [ ] 准备 rollback 方案（回滚到上一 tag + DB migration 兼容性确认）

### Migration 测试

- [ ] 在测试环境执行 AutoMigrate，确认无报错
- [ ] 验证 migration 可重复执行（GORM AutoMigrate 幂等）
- [ ] 确认无锁表时间过长的 migration

## 部署后

### 健康检查

- [ ] health check 通过（GET /api/status 返回 success:true）
- [ ] 容器状态正常（docker ps 无重启循环）

### 功能验证

- [ ] 登录测试（admin 账户可登录）
- [ ] Token 创建测试（创建 / 删除 / 列表）
- [ ] 充值 webhook 测试（Stripe/Creem/Epay 回调签名验证通过）
- [ ] 消耗额度测试（发起一次 API 请求，验证 quota 扣减 + 日志记录）
- [ ] BatchUpdate 检查（BATCH_UPDATE_ENABLED=true 时，5s 后验证 DB 有增量更新）

### 监控报警

- [ ] 日志报警检查（确认 SysError 含 manual_intervention_required=true 的日志可被采集告警）
- [ ] critical 场景验证（模拟 DB 故障，确认 BatchUpdate 降级告警触发）
- [ ] shutdown 验证（docker-compose down，确认 FlushBatchUpdate 日志输出 + 无数据丢失）

### 日志验证

- [ ] 业务日志正常输出（/app/logs 目录有日志文件）
- [ ] 错误日志可追溯（包含 user_id / token_id / channel_id 等结构化字段）
- [ ] 审计日志完整（管理操作有 admin_info 审计字段）
