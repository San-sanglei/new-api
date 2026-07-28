package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// migrateTenantPhase2 第二阶段 A：业务表增加 tenant_id 字段（幂等）
//
// 迁移内容：
//  1. tokens.tenant_id（INT DEFAULT 1）
//  2. logs.tenant_id（INT DEFAULT 1）
//  3. 历史数据回填为 DefaultTenantID（id=1）
//
// 注意：
//   - logs 表可能使用独立 LOG_DB（LOG_SQL_DSN 配置时），需分别处理；
//     当 LOG_SQL_DSN 未设置时 LOG_DB == DB，此时 logs 与 tokens 同库。
//   - 本阶段只加字段 + 回填历史数据，不修改任何查询逻辑（控制器/路由不变）。
//   - Phase 2-B 才会修改查询，加入 tenant_id 过滤。
//   - 不修改支付/Epay，不修改 controller/router。
func migrateTenantPhase2() error {
	if DB == nil {
		return errors.New("database not initialized")
	}

	// 1. tokens.tenant_id（主 DB）
	if err := addColumnIfNotExistsOnDb(DB, "tokens", "tenant_id", "INT DEFAULT 1"); err != nil {
		return fmt.Errorf("add tenant_id to tokens failed: %w", err)
	}

	// 2. logs.tenant_id（LOG_DB；当 LOG_SQL_DSN 未设置时 LOG_DB == DB）
	if LOG_DB != nil {
		if err := addColumnIfNotExistsOnDb(LOG_DB, "logs", "tenant_id", "INT DEFAULT 1"); err != nil {
			return fmt.Errorf("add tenant_id to logs failed: %w", err)
		}
	}

	// 3. 回填历史数据：tokens
	if err := backfillTenantIdOnDb(DB, "tokens"); err != nil {
		return fmt.Errorf("backfill tokens.tenant_id failed: %w", err)
	}

	// 4. 回填历史数据：logs
	if LOG_DB != nil {
		if err := backfillTenantIdOnDb(LOG_DB, "logs"); err != nil {
			return fmt.Errorf("backfill logs.tenant_id failed: %w", err)
		}
	}

	common.SysLog("tenant phase2-A migration completed")
	return nil
}

// addColumnIfNotExistsOnDb 在指定 gorm.DB 上幂等添加列（跨 SQLite/MySQL/PostgreSQL）
//
// 与 addColumnIfNotExists（操作主 DB）的区别：本函数接受显式 db 参数，
// 用于支持 LOG_DB 等独立数据库实例的列添加。检测逻辑与 addColumnIfNotExists 一致，
// 通过 Dialector.Name() 判断数据库类型，避免依赖 common.UsingPostgreSQL 等全局标志
// （这些标志针对主 DB，LOG_DB 可能使用不同数据库类型）。
func addColumnIfNotExistsOnDb(db *gorm.DB, table, column, definition string) error {
	if db == nil {
		return errors.New("database not initialized")
	}

	dialect := db.Dialector.Name()
	var count int64
	var err error
	switch dialect {
	case "postgres":
		err = db.Raw(
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_name = ? AND column_name = ?`,
			table, column,
		).Count(&count).Error
	case "mysql":
		err = db.Raw(
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
			table, column,
		).Count(&count).Error
	default: // sqlite
		err = db.Raw(
			`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
			table, column,
		).Count(&count).Error
	}
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 列已存在
	}

	// 注意：column/definition 来自代码内常量（非用户输入），可安全拼接。
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	if err := db.Exec(sql).Error; err != nil {
		return err
	}
	common.SysLog(fmt.Sprintf("added column %s.%s", table, column))
	return nil
}

// backfillTenantIdOnDb 回填历史数据的 tenant_id 为 DefaultTenantID
//
// 兜底处理：
//   - 若 ALTER TABLE ADD COLUMN 时 DEFAULT 1 已对历史行生效，则 RowsAffected=0，跳过。
//   - 若部分数据库版本未将 DEFAULT 应用于历史行（如 SQLite 部分场景），
//     则将 0 或 NULL 的行更新为 DefaultTenantID。
func backfillTenantIdOnDb(db *gorm.DB, table string) error {
	if db == nil {
		return errors.New("database not initialized")
	}

	result := db.Table(table).
		Where("tenant_id = ? OR tenant_id IS NULL", 0).
		Update("tenant_id", DefaultTenantID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		common.SysLog(fmt.Sprintf("backfilled %d rows in %s.tenant_id to default tenant", result.RowsAffected, table))
	}
	return nil
}
