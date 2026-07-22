# Release Checklist - Race Test CI

## Linux CI 执行

在生产发布前，CI/CD 必须在 Linux 环境执行以下命令进行 race detector 测试：

```bash
CGO_ENABLED=1 go test -race -timeout 600s ./model/... ./service/...
```

## Windows 本地无法执行

Windows 环境无 gcc，CGO 不可用，`go test -race` 会报错：

```
race: enabled, but CGO is not enabled
```

## 替代方案

1. 本地开发使用普通 `go test ./...` 验证逻辑
2. race detector 必须在 Linux CI 环境执行
3. 重点覆盖路径：
   - `./model/...` — BatchUpdate snapshot swap、quota CAS、retry counter
   - `./service/...` — quota 并发扣减、PostConsumeQuota rollback

## 验证标准

- `go test -race` 必须 PASS，无 race 报告
- 若发现 race，必须修复后重新发布
