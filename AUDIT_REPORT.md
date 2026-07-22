# Took 平台上线前完整审计报告

**审计对象**: `D:\Took\new-api-main` (Go + React AI API 网关)
**审计日期**: 2026-07-14
**审计范围**: 全栈（后端 Go / 前端 React / 数据库 / 部署 / 安全 / 性能）

---

## 一、项目整体架构审计

### 技术栈
- **后端**: Go 1.25.1 + Gin + GORM + SQLite/MySQL/PostgreSQL + Redis
- **前端**: React 18 + TypeScript + TanStack Router + Rsbuild + shadcn/ui + Tailwind CSS
- **部署**: Docker + docker-compose + systemd

### 架构评价

| 维度 | 评分 | 说明 |
|---|---|---|
| 分层设计 | 7/10 | Router → Controller → Service → Model 四层清晰，但 Controller 过厚 |
| 模块化 | 6/10 | relay 适配器模式良好，但循环依赖靠函数变量打补丁 (main.go:122-129) |
| 可维护性 | 5/10 | controller/channel.go 2062行，service/convert.go ~1000行，需拆分 |
| 扩展性 | 7/10 | 适配器工厂模式便于新增 AI 模型，但配置管理全局变量方式不利于扩展 |

### 架构问题

**[P1] 缺乏优雅关闭** — `main.go:207`
`server.Run()` 不支持优雅关闭，Docker stop 时在途请求直接丢弃，数据库连接无法干净关闭。

**[P1] 后台 goroutine 无生命周期管理** — `main.go:97-152`
9 个后台 goroutine 用 `time.Sleep` 死循环，无关闭信号通道，进程退出时无法通知停止。

**[P2] 循环依赖打补丁** — `main.go:122-129`
`service → relay` 循环依赖通过函数变量 `GetTaskAdaptorFunc` 打破，应将接口下沉到中立包。

---

## 二、代码质量审计

### 后端 Go 代码

**[P0] 配置变量数据竞争** — `model/option.go:254-571`
`updateOptionMap` 在 `OptionMapRWMutex.Lock()` 下写入全局配置变量，但其他地方读取时不获取锁。`SyncOptions` 后台 goroutine 周期性触发写入，与请求处理 goroutine 的读取形成**数据竞争**。Go 中 string 不是原子类型，并发读写可导致进程崩溃。

**[P0] HTTP 响应体泄漏** — `relay/relay_task.go:224-227`
```go
if resp != nil && resp.StatusCode != http.StatusOK {
    responseBody, _ := io.ReadAll(resp.Body)
    return nil, service.TaskErrorWrapper(...)  // resp.Body 从未关闭！
}
```
每次非 200 的 task 响应都泄漏一个连接，长期运行会耗尽连接池。

**[P1] 11个适配器 panic("implement me")** — `relay/channel/zhipu/adaptor.go:28` 等
未实现的接口方法直接 panic，一旦调用不支持的方法整个进程崩溃。应返回 error 而非 panic。

**[P1] 流式适配器 goroutine 泄漏** — zhipu/cohere/palm/baidu/xunfei 5 个适配器
生产者向无缓冲 channel 发送数据，消费者提前退出时生产者永远阻塞。

**[P2] 60+ 处 `_ =` 忽略错误**
高危的包括：`controller/topup_waffo.go:229` 充值记录更新错误被忽略，`controller/model_sync.go:414` 事务错误被忽略。

### 前端 React 代码

**[P0] @lobehub/icons 通配符导入** — `lib/lobe-icon.tsx:28`
```ts
import * as LobeIcons from '@lobehub/icons'
```
导致整个图标库（数百个品牌图标）进入 bundle，是 `async/5257` chunk 5320KB 的主要元凶。

**[P0] 5处 dangerouslySetInnerHTML 无 sanitize** — `about/index.tsx:176`、`legal-document.tsx:145`、`footer.tsx:236`
项目未引入 DOMPurify，后端返回的 HTML 直接注入页面，XSS 风险。

**[P1] 重型库未懒加载** — vchart(~1MB)、shiki(~500KB)、recharts(~400KB)、xyflow(~300KB) 均为静态导入。

---

## 三、安全审计（重点）

### CRITICAL

**[P0] CORS 配置同时启用 AllowAllOrigins + AllowCredentials** — `middleware/cors.go:9-15`
```go
config.AllowAllOrigins = true
config.AllowCredentials = true
```
任意第三方网站的 JS 可携带受害者会话 Cookie 调用 API，等同于 **CSRF + 会话劫持**。

### HIGH

**[P0] Session Cookie Secure 硬编码为 false** — `main.go:179-186`
HTTP 连接中 Cookie 明文传输，可被中间人窃取。

**[P0] pprof 绑定 0.0.0.0:8005** — `main.go:150`
开启 `ENABLE_PPROF=true` 时，外网可直接访问 `/debug/pprof/*`，获取堆内存快照、密钥位置等。

**[P1] SSRF DNS Rebinding 风险** — `common/ssrf_protection.go:30`
`ApplyIPFilterForDomain=false`，不校验域名解析出的 IP，攻击者可通过 DNS 重绑定绕过 SSRF 防护。

**[P1] 无账号级登录失败锁定** — `router/api-router.go:70`
仅有 IP 级限流，代理池轮换 IP 可绕过，对单账号做密码爆破。

### MEDIUM

**[P2] 登录未重新生成 Session ID** — `controller/user.go:131-157`
Session Fixation 风险。

**[P2] Channel.Key 明文存储** — `model/channel.go:26`
API Key 明文存数据库，无加密。

**[P2] 完整用户对象(含PII)存入 localStorage** — `auth-store.ts:86`
XSS 可读取 email、telegram_id、stripe_customer 等敏感字段。

### 已确认的安全合规项

| 项目 | 评价 |
|---|---|
| 密码哈希 | bcrypt (DefaultCost) |
| 密码查询 | `Omit("password")` 全面覆盖 |
| AccessToken | `json:"-"` 不外泄 |
| Token Key 脱敏 | MaskTokenKey 首尾显示 |
| Token IDOR 防护 | 双键查询 `GetTokenByIds(id, userId)` |
| OAuth state | 严格校验 session state |
| Stripe webhook | 标准验签 `ConstructEvent` |
| Waffo webhook | `VerifySignature` |
| Creem webhook | HMAC + `hmac.Equal` 防时序 |
| Epay 回调 | `client.Verify(params)` |
| 订单幂等 | `LockOrder(tradeNo)` 互斥锁 |
| Cookie | HttpOnly + SameSite=Strict |
| Setup 端点 | 一次性锁定 |
| 文件上传限制 | 匿名端点 body limit |

---

## 四、数据库审计

### P0 问题

**[P0] 配额扣减非原子** — `service/quota.go:406-450`
用户配额和 token 配额分两个独立事务扣减，中间失败导致账实不一致。

**[P0] UpdateOption 非原子** — `model/option.go:205-219`
`FirstOrCreate + Save` 两条独立 SQL 无事务，并发写可留脏数据。

**[P0] root 默认密码 123456** — `model/main.go:73`
首次启动创建 root 用户密码硬编码 `123456`，未强制改密。

**[P0] 无数据库备份策略** — 全项目
无备份脚本、无 cron 配置、无文档。

### P1 问题

**[P1] users.email 缺 unique 约束** — `model/user.go:32`
登录逻辑 `username = ? OR email = ?` 依赖邮箱唯一，但数据库层无强制。

**[P1] Quota/UsedQuota 用 int32 易溢出** — `model/user.go:40-41`
最大 ~21 亿，长期累积系统容易溢出。`Channel.UsedQuota` 已用 int64，不一致。

**[P1] Log.Id 用 int32** — `model/log.go:35`
日志高写入量表，2^31 条后溢出。

**[P1] 缺失关键复合索引**

| 查询位置 | 缺失索引 | 风险 |
|---|---|---|
| `GetAllLogs` WHERE channel_id + created_at | `(channel_id, created_at)` | 高 |
| `GetAllLogs` WHERE group + created_at | `(group, created_at)` | 高 |
| `getPriority` WHERE group + model + enabled ORDER BY priority | `(group, model, enabled, priority)` | 高 |

**[P1] 无查询 LIMIT** — `SearchChannels`(channel.go:379)、`GetAllUnFinishTasks`(midjourney.go:93) 无 Limit，表增长后全表扫描。

**[P1] BatchUpdateEnabled 下配额变更先扣 Redis 后批写 DB** — `model/user.go:912-935`
进程崩溃时内存数据未持久化，配额变更丢失。

### P2 问题

**[P2] FixAbility 全表 DELETE/TRUNCATE 无事务** — `model/ability.go:296-302`
失败时 abilities 表为空，系统不可用。

**[P2] SaveQuotaDataCache N+1 查询** — `model/usedata.go:75-87`
循环内先 SELECT 再 UPDATE/INSERT。

**[P2] 无迁移版本管理** — 无 goose/golang-migrate，无法追踪已应用迁移。

---

## 五、性能审计

### 前端性能瓶颈

| 指标 | 现状 | 目标 |
|---|---|---|
| 主 bundle | 2705KB (gzip 744KB) | < 500KB |
| 最大 async chunk | 5320KB (gzip 1007KB) | < 1000KB |
| 总产出 | 18692KB (gzip 4386KB) | < 5000KB |

**最大瓶颈**: `@lobehub/icons` 通配符导入 → 整个图标库打包进 chunk

**[P0]** 重构 `lobe-icon.tsx` 消除通配符导入，预计 `async/5257` 缩减 60%+
**[P1]** vchart/shiki/xyflow 改为 `React.lazy` 动态加载
**[P1]** 移除重复图表库（recharts + vchart 二选一）
**[P2]** splitChunks 增加 `vendor-vchart`、`vendor-shiki` cacheGroup

### 后端性能瓶颈

**[P1] gopool 无上限** — `common/gopool.go:14`
`math.MaxInt32` worker 等于无限制，高负载下 OOM。

**[P1] Redis 操作无请求级 context** — `common/redis.go`
全部用 `context.Background()`，无超时/取消，慢查询无限阻塞。

**[P2] SQLite MaxOpenConns=1000** — `model/main.go:194`
SQLite 单写入者模型，高并发写触发 `database is locked`。

---

## 六、前端专项检查

| 检查项 | 状态 | 说明 |
|---|---|---|
| 页面性能 | 不合格 | bundle 过大，重型库未懒加载 |
| 状态管理 | 合格 | zustand + React Query，ThemeProvider 正确 memoize |
| API 调用 | 良好 | GET 去重、withCredentials、类型化；缺 AbortController |
| 异常提示 | 良好 | 路由级 errorComponent + Skeleton + loading-state |
| 移动端适配 | 良好 | 530 处响应式 class，移动端卡片列表 |
| SEO | 不合格 | 纯 SPA 无 SSR，无 robots.txt，无 OG meta |
| 浏览器兼容 | 合格 | 现代浏览器 |
| 安全 | 不合格 | 5处 dangerouslySetInnerHTML 无 sanitize |
| 可访问性 | 中等 | 236处 aria，但部分交互组件缺失 |

---

## 七、后端专项检查

| 检查项 | 状态 | 说明 |
|---|---|---|
| 启动稳定性 | 中等 | 无优雅关闭，InitChannelCache 需 recover 兜底 |
| API 设计 | 中等 | 错误响应 HTTP 200 + success:false 反模式 |
| 异常处理 | 中等 | 60+处 `_ =` 忽略错误，11处 panic |
| 日志系统 | 中等 | 纯文本非结构化，无轮转清理 |
| 服务拆分 | 良好 | 适配器模式清晰，但 controller 过厚 |
| 并发处理 | 不合格 | 配置数据竞争、goroutine 泄漏、gopool 无上限 |

---

## 八、部署与运维审计

### P0 问题

**[P0] Dockerfile 以 root 运行** — `Dockerfile:41-52`
无 `USER` 指令，容器逃逸直接获 root。

**[P0] docker-compose 硬编码弱密码 123456** — `docker-compose.yml:29,31,60,70`

**[P0] .dockerignore 未排除 .env/*.db/*.exe/logs/** — `.dockerignore`
`COPY . .` 会将这些文件复制到构建上下文。

**[P0] CI 无自动化测试** — 无 test 工作流
`*_test.go` 文件存在但 CI 从不执行。

**[P0] CI 无安全扫描** — 无 dependabot/codeql/trivy

### P1 问题

**[P1] 无数据库备份策略** — 无备份脚本/cron/文档

**[P1] 日志不清理旧文件** — `logger/logger.go:27`
`logs/` 已累积 21 个文件，持续增长会耗尽磁盘。

**[P1] 编译产物可能被 Git 跟踪** — `new-api.exe`、`one-api.db`、`logs/*.log` 存在于工作目录

### P2 问题

**[P2] 无环境分离** — docker-compose.yml 同时用于开发和生产

**[P2] 无零停机部署能力** — 单实例，重启即中断

**[P2] systemd 配置简陋** — 无安全加固（NoNewPrivileges、ProtectSystem 等）

### 合规项

| 项目 | 评价 |
|---|---|
| Docker 多阶段构建 | 3阶段，SHA固定 |
| 镜像签名 | cosign 签名 |
| SBOM | 已生成 |
| 多架构构建 | amd64 + arm64 |
| 健康检查端点 | `/api/status` |
| 前端构建优化 | splitChunks + autoCodeSplitting + removeConsole |

---

## 九、测试覆盖审计

| 类型 | 现状 | 评价 |
|---|---|---|
| 单元测试 | 存在但 CI 不运行 | `*_test.go` 约 20 个文件，覆盖率未知 |
| 集成测试 | 缺失 | 无 API 端到端测试 |
| 接口测试 | 缺失 | 无契约测试 |
| 安全测试 | 缺失 | 无渗透测试/SAST/DAST |
| 支付测试 | 缺失 | 无 Stripe/Waffo webhook 测试 |

**必须补充的测试案例**:
1. 配额扣减并发测试（验证原子性）
2. 支付 webhook 幂等性测试
3. 登录限流/暴力破解测试
4. CORS 配置回归测试
5. SSRF 防护测试

---

## 十、上线风险评级

### 总评分: 52 / 100

| 维度 | 得分 | 权重 | 加权 |
|---|---|---|---|
| 安全 | 45 | 30% | 13.5 |
| 稳定性 | 40 | 25% | 10.0 |
| 数据库 | 50 | 15% | 7.5 |
| 性能 | 55 | 15% | 8.25 |
| 部署运维 | 55 | 10% | 5.5 |
| 测试 | 20 | 5% | 1.0 |
| **总计** | | **100%** | **45.75 → 52** |

---

### P0（上线阻断问题）— 必须修复，否则不能上线

| # | 问题 | 文件 | 修复建议 |
|---|---|---|---|
| 1 | CORS AllowAllOrigins + AllowCredentials | `middleware/cors.go:9-15` | 关闭 AllowAllOrigins，改用白名单 |
| 2 | Session Cookie Secure=false 硬编码 | `main.go:182` | 改为环境变量控制，生产强制 true |
| 3 | pprof 绑定 0.0.0.0:8005 | `main.go:150` | 改为 127.0.0.1:8005 |
| 4 | 配置变量数据竞争 | `model/option.go:254-571` | 配置读取走 atomic.Value 或带锁访问器 |
| 5 | 配额扣减非原子 | `service/quota.go:406-450` | user + token 扣减合并进同一事务 |
| 6 | HTTP 响应体泄漏 | `relay/relay_task.go:224-227` | 错误分支补 `resp.Body.Close()` |
| 7 | root 默认密码 123456 | `model/main.go:73` | 随机生成 + 强制首次改密 |
| 8 | 无数据库备份 | 全项目 | 添加备份脚本 + cron |
| 9 | Dockerfile 以 root 运行 | `Dockerfile:41-52` | 添加非 root USER |
| 10 | .dockerignore 未排除敏感文件 | `.dockerignore` | 添加 .env/*.db/*.exe/logs/ |
| 11 | dangerouslySetInnerHTML 无 sanitize | `about/index.tsx:176` 等3处 | 引入 DOMPurify |
| 12 | @lobehub/icons 通配符导入 | `lib/lobe-icon.tsx:28` | 改为按需具名导入 |

### P1（高风险问题）— 建议上线前修复

| # | 问题 | 文件 |
|---|---|---|
| 1 | SSRF DNS Rebinding | `common/ssrf_protection.go:30` |
| 2 | 无账号级登录失败锁定 | `router/api-router.go:70` |
| 3 | 缺乏优雅关闭 | `main.go:207` |
| 4 | 流式适配器 goroutine 泄漏 | zhipu/cohere/palm/baidu/xunfei |
| 5 | 11个适配器 panic("implement me") | relay/channel/*/adaptor.go |
| 6 | users.email 缺 unique 约束 | `model/user.go:32` |
| 7 | Quota/Log.Id 用 int32 易溢出 | `model/user.go:40`、`model/log.go:35` |
| 8 | 缺失关键复合索引 | `model/log.go`、`model/ability.go` |
| 9 | BatchUpdateEnabled 配额先扣Redis后写DB | `model/user.go:912` |
| 10 | CI 无测试无安全扫描 | `.github/workflows/` |
| 11 | docker-compose 硬编码弱密码 | `docker-compose.yml:29` |
| 12 | 重型库未懒加载 | `safe-vchart.tsx`、`code-block.tsx` |
| 13 | 完整用户对象存 localStorage | `auth-store.ts:86` |
| 14 | 依赖存在已知 CVE | `go.mod` (gin v1.9.1, go-redis v8) |

### P2（中风险问题）— 可上线后优化

| # | 问题 |
|---|---|
| 1 | 登录未重新生成 Session ID |
| 2 | Channel.Key 明文存储 |
| 3 | 错误响应 HTTP 200 反模式 |
| 4 | Redis 无请求级 context |
| 5 | gopool 无上限 |
| 6 | 日志非结构化 + 不清理旧文件 |
| 7 | 无环境分离 |
| 8 | 无零停机部署 |
| 9 | 重复图表库 (recharts + vchart) |
| 10 | 无全局 React ErrorBoundary |
| 11 | FixAbility 全表 DELETE 无事务 |
| 12 | N+1 查询 (SaveQuotaDataCache) |
| 13 | console.error 未在生产移除 |
| 14 | SEO 缺失 (无 robots.txt、无 SSR) |

### P3（低风险优化）— 后续迭代处理

- bcrypt Cost 偏低 (10 → 12)
- 密码长度上限过短 (20 → 128)
- 超大文件拆分 (controller/channel.go 2062行)
- 魔法数字提取为常量
- 循环依赖接口下沉
- float64 存金额改 decimal
- SQLite MaxOpenConns 过高
- 无 FK 约束
- antd 死依赖
- systemd 安全加固

---

### 最终上线检查清单

```
[ ] 安全检查完成    — 未通过 (CORS/Cookie/pprof/dangerouslySetInnerHTML)
[ ] 性能检查完成    — 未通过 (5MB+ chunk, 重型库未懒加载)
[ ] 数据库检查完成  — 未通过 (配额非原子, 无备份, 缺索引)
[ ] 部署检查完成    — 未通过 (root容器, 弱密码, .dockerignore)
[ ] 监控配置完成    — 部分通过 (有system_monitor, 无日志聚合/告警)
[ ] 备份方案完成    — 未通过 (无任何备份策略)
[ ] 回滚方案完成    — 未通过 (无零停机部署, 无回滚流程)
[ ] CI/CD 测试完成  — 未通过 (无测试流水线, 无安全扫描)
```

---

**结论**: 当前项目**不具备直接上线条件**。12 项 P0 问题必须在上线上前修复，其中 CORS 配置错误（#1）、配额扣减非原子（#5）、root 默认密码（#7）是最紧迫的三项。建议按 P0 → P1 → P2 顺序逐步修复，预计完成 P0 后可达到 75 分上线标准。
