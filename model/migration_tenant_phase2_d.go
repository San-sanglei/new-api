package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
)

// migrateTenantPhase2D 第二阶段 D：剩余业务表增加 tenant_id 字段（幂等）
//
// 迁移内容：
//  1. subscription_plans / subscription_orders / user_subscriptions / subscription_pre_consume_records
//     加 tenant_id（INT DEFAULT 1）
//  2. top_ups / redemptions / tasks / midjourneys 加 tenant_id（INT DEFAULT 1）
//  3. 历史数据回填为 DefaultTenantID（id=1）
//
// 注意：
//   - 仅加字段 + 回填历史数据，不修改任何业务逻辑。
//   - 不修改 Channel / Ability / Pricing / Vendor / Group。
//   - 不修改支付/Epay 流程，不修改 Relay channel 选择逻辑，不修改 Role 权限模型。
//   - 查询/写入路径的 tenant_id 过滤由 controller/model 层修改实现（本迁移只负责 DDL）。
func migrateTenantPhase2D() error {
	if DB == nil {
		return errors.New("database not initialized")
	}

	tables := []string{
		"subscription_plans",
		"subscription_orders",
		"user_subscriptions",
		"subscription_pre_consume_records",
		"top_ups",
		"redemptions",
		"tasks",
		"midjourneys",
	}

	for _, table := range tables {
		if err := addColumnIfNotExistsOnDb(DB, table, "tenant_id", "INT DEFAULT 1"); err != nil {
			return fmt.Errorf("add tenant_id to %s failed: %w", table, err)
		}
		if err := backfillTenantIdOnDb(DB, table); err != nil {
			return fmt.Errorf("backfill %s.tenant_id failed: %w", table, err)
		}
	}

	common.SysLog("tenant phase2-D migration completed")
	return nil
}
