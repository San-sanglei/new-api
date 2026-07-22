# New-API 免费架构部署指南

## 架构总览

```
用户浏览器 → took.ink (Vercel 前端)
                 ↓ /api/* /mj/* /pg/* (Vercel Rewrite 代理)
            Render Go 后端
                 ↓
            ┌────┴────┐
        Supabase     Upstash
        PostgreSQL   Redis
```

| 组件 | 平台 | 用途 | 免费额度 |
|------|------|------|----------|
| 前端 | Vercel | React 静态站 + API 代理 | 100GB/月带宽 |
| 后端 | Render | Go API 长驻服务 | 750h/月, 512MB RAM |
| 数据库 | Supabase | PostgreSQL | 500MB 存储, 2 个项目 |
| 缓存 | Upstash | Redis | 10K 命令/天 |

---

## 步骤 1：代码仓库已推送到 Gitee

代码已推送至：`https://gitee.com/San-sang/new-api`

后续步骤直接使用此 Gitee 仓库地址导入到 Render 和 Vercel。

---

## 步骤 2：创建 Supabase 数据库

1. 打开 https://supabase.com → 注册/登录
2. New Project → 填写名称 `new-api-db`
3. 设置数据库密码（记下来）
4. 等待项目创建完成（约 2 分钟）
5. 进入 Settings → Database → Connection string
6. 复制 **URI** 格式连接串，格式如下：

```
postgresql://postgres:[YOUR_PASSWORD]@db.[PROJECT_REF].supabase.co:5432/postgres
```

将 `[YOUR_PASSWORD]` 替换为你设置的密码。这就是 `SQL_DSN` 的值。

---

## 步骤 3：创建 Upstash Redis

1. 打开 https://upstash.com → 注册/登录
2. Create Database → 名称 `new-api-redis` → region 选离 Supabase 最近的区域
3. 创建完成后复制 **Redis URL**，格式如下：

```
redis://default:[REDIS_PASSWORD]@global-[xxx].upstash.io:6379
```

这就是 `REDIS_CONN_STRING` 的值。

---

## 步骤 4：部署 Go 后端到 Render

1. 打开 https://render.com → 注册/登录（可用 GitHub 或 GitLab 登录）
2. New + → Blueprint
3. 选择 **Public Git Repository**，填入 Gitee 仓库地址：`https://gitee.com/San-sang/new-api`
   > 注：如 Render 不支持 Gitee 直连，可在 Render 网页上手动上传 `Dockerfile` 所在仓库镜像，或先在 GitHub 镜像一份代码。
4. Render 会自动识别 `render.yaml`，显示配置摘要
5. 在环境变量中填入：

| 变量名 | 值 |
|--------|-----|
| `SQL_DSN` | 步骤 2 的 Supabase 连接串 |
| `REDIS_CONN_STRING` | 步骤 3 的 Upstash Redis URL |
| `CORS_ALLOWED_ORIGINS` | 先留空，部署 Vercel 后填入 |

6. 点击 **Apply** 开始构建
7. 构建完成后，Render 会分配一个 URL，如：`https://new-api-backend.onrender.com`
8. 验证：访问 `https://new-api-backend.onrender.com/api/status`，应返回 `{"success":true,...}`

**记下这个 Render URL，下一步需要。**

---

## 步骤 5：部署前端到 Vercel

1. 打开 https://vercel.com → 注册/登录（用 GitHub 或 GitLab 账号）
2. Add New → Project
3. 选择 **Import Git Repository** → 填入 Gitee 仓库地址：`https://gitee.com/San-sang/new-api`
   > 注：Vercel 原生支持 GitHub/GitLab/Bitbucket。Gitee 仓库需通过 Vercel 的"Import from URL"功能导入；如不可用，建议用 `vercel` CLI 从本地目录直接部署（见下方备选方案）。
4. 在配置页面设置：

| 配置项 | 值 |
|--------|-----|
| Root Directory | `web/default` |
| Framework Preset | Other |
| Build Command | `bun install && bun run build` |
| Output Directory | `dist` |
| Install Command | 留空（Vercel 自动检测 Bun） |

5. 展开 **Environment Variables**，添加：

| 变量名 | 值 |
|--------|-----|
| `API_URL` | `https://new-api-backend.onrender.com`（步骤 4 的 Render URL，**不带末尾斜杠**） |

6. 点击 **Deploy**
7. 等待构建完成（约 3-5 分钟）
8. Vercel 会分配 URL，如：`https://new-api-xxx.vercel.app`

**验证**：访问该 URL，应看到登录页面。

### 备选方案：用 Vercel CLI 从本地部署（推荐，无需 Gitee 关联）

由于 Vercel 对 Gitee 非原生支持，推荐用 CLI 直接从本地推送构建产物：

```powershell
# 1. 安装 Vercel CLI
npm i -g vercel

# 2. 进入前端目录
cd d:\Took\new-api-main\web\default

# 3. 登录（会打开浏览器）
vercel login

# 4. 部署（首次会问几个问题，按下方回答）
#    - Set up and deploy?  Y
#    - Which scope?        选你的账号
#    - Link to existing project?  N
#    - Project name?       new-api
#    - Directory?          current (./)
vercel

# 5. 部署到生产
vercel --prod

# 6. 设置环境变量（替换为 Render 后端 URL）
vercel env add API_URL production
# 粘贴值: https://new-api-backend.onrender.com

# 7. 重新部署使环境变量生效
vercel --prod
```

---

## 步骤 6：回填 CORS 配置

回到 Render Dashboard → 环境变量：

将 `CORS_ALLOWED_ORIGINS` 设置为 Vercel 分配的域名（及最终的自定义域名）：

```
https://new-api-xxx.vercel.app,https://took.ink
```

保存后 Render 会自动重新部署。

---

## 步骤 7：绑定 took.ink 域名

### 7.1 在 Vercel 添加自定义域名

1. Vercel Dashboard → 你的项目 → Settings → Domains
2. 输入 `took.ink` → Add
3. Vercel 会显示需要添加的 DNS 记录

### 7.2 在阿里云 DNS 添加解析

登录阿里云域名控制台 https://dc.console.aliyun.com → DNS 修改：

| 记录类型 | 主机记录 | 记录值 |
|----------|---------|--------|
| CNAME | `@` | `cname.vercel-dns.com` |
| CNAME | `www` | `cname.vercel-dns.com` |

或如果 Vercel 要求 A 记录：

| 记录类型 | 主机记录 | 记录值 |
|----------|---------|--------|
| A | `@` | `76.76.21.21` |
| CNAME | `www` | `cname.vercel-dns.com` |

### 7.3 等待 DNS 生效

DNS 生效通常需要 10 分钟 ~ 2 小时。Vercel 会自动检测并签发 SSL 证书。

---

## 环境变量完整清单

### Vercel（前端）

| 变量名 | 说明 |
|--------|------|
| `API_URL` | Render 后端地址，如 `https://new-api-backend.onrender.com` |

### Render（后端）

| 变量名 | 说明 |
|--------|------|
| `SQL_DSN` | Supabase PostgreSQL 连接串 |
| `REDIS_CONN_STRING` | Upstash Redis 连接串 |
| `SESSION_SECRET` | 会话密钥（Render 自动生成） |
| `CRYPTO_SECRET` | 加密密钥（Render 自动生成） |
| `CORS_ALLOWED_ORIGINS` | 允许的前端域名，逗号分隔 |
| `GIN_MODE` | `release` |
| `PORT` | `3000` |
| `TZ` | `Asia/Shanghai` |

---

## 常见问题

### Q: Render 免费版会休眠吗？
A: 是，15 分钟无请求后休眠，首次唤醒需 30-60 秒。生产建议升级到 Starter ($7/月)。

### Q: AI 流式响应会被 Vercel 代理截断吗？
A: 不会。Vercel Rewrites 是透明代理，不截断长连接。

### Q: 首次启动报错 "database connection failed"？
A: 检查 `SQL_DSN` 密码是否正确，Supabase 项目是否 active。Supabase 免费版会暂停一周不活动的项目。

### Q: 前端构建失败 "bun not found"？
A: Vercel 会自动安装 Bun。如失败，将 Build Command 改为 `npm install && npx rsbuild build`。
