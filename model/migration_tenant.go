package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
)

// migrateTenantSchema 执行第一阶段多租户迁移（幂等）
//
// 迁移内容：
//  1. AutoMigrate tenants / tenant_members 表
//  2. 创建默认租户（id=1, Took Official）
//  3. users 表增加 current_tenant_id 列（默认 1）
//  4. 为现有用户创建 tenant_members 记录（默认租户的成员）
//
// 注意：本迁移不修改任何业务表（tokens/logs/top_ups 等），不修改支付逻辑。
// 第二阶段起才会给业务表加 tenant_id 字段。
func migrateTenantSchema() error {
	if DB == nil {
		return errors.New("database not initialized")
	}

	// 1. AutoMigrate Tenant 和 TenantMember 表
	if err := DB.AutoMigrate(&Tenant{}, &TenantMember{}); err != nil {
		return fmt.Errorf("auto-migrate tenant tables failed: %w", err)
	}

	// 2. 确保默认租户存在
	if err := EnsureDefaultTenant(); err != nil {
		return fmt.Errorf("ensure default tenant failed: %w", err)
	}

	// 3. users 表加 current_tenant_id 列（幂等）
	if err := addColumnIfNotExists("users", "current_tenant_id", "INT DEFAULT 1"); err != nil {
		return fmt.Errorf("add current_tenant_id to users failed: %w", err)
	}

	// 4. 为现有用户创建 tenant_members 记录（迁移到默认租户，幂等）
	if err := migrateExistingUsersToDefaultTenant(); err != nil {
		return fmt.Errorf("migrate existing users to default tenant failed: %w", err)
	}

	common.SysLog("tenant schema migration completed")
	return nil
}

// addColumnIfNotExists 幂等添加列（跨 SQLite/MySQL/PostgreSQL）
func addColumnIfNotExists(table, column, definition string) error {
	if DB == nil {
		return errors.New("database not initialized")
	}

	var count int64
	switch {
	case common.UsingPostgreSQL:
		if err := DB.Raw(
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_name = ? AND column_name = ?`,
			table, column,
		).Count(&count).Error; err != nil {
			return err
		}
	case common.UsingMySQL:
		if err := DB.Raw(
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
			table, column,
		).Count(&count).Error; err != nil {
			return err
		}
	case common.UsingSQLite:
		if err := DB.Raw(
			`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
			table, column,
		).Count(&count).Error; err != nil {
			return err
		}
	default:
		// 兜底：尝试 SQLite 语法
		if err := DB.Raw(
			`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
			table, column,
		).Count(&count).Error; err != nil {
			return err
		}
	}

	if count > 0 {
		return nil // 列已存在
	}

	// 注意：使用占位符传递列名和定义在不同数据库间不通用，
	// 这里 column/definition 来自代码内常量（非用户输入），可安全拼接。
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	if err := DB.Exec(sql).Error; err != nil {
		return err
	}
	common.SysLog(fmt.Sprintf("added column %s.%s", table, column))
	return nil
}

// migrateExistingUsersToDefaultTenant 将未在默认租户中的用户加入默认租户（幂等）
func migrateExistingUsersToDefaultTenant() error {
	if DB == nil {
		return errors.New("database not initialized")
	}

	// 查找未在默认租户中的用户 ID
	var userIds []int
	err := DB.Table("users u").
		Joins("LEFT JOIN tenant_members tm ON tm.user_id = u.id AND tm.tenant_id = ?", DefaultTenantID).
		Where("tm.id IS NULL").
		Pluck("u.id", &userIds).Error
	if err != nil {
		return err
	}

	if len(userIds) == 0 {
		return nil
	}

	// 批量插入 tenant_members
	members := make([]TenantMember, 0, len(userIds))
	for _, uid := range userIds {
		members = append(members, TenantMember{
			TenantId: DefaultTenantID,
			UserId:   uid,
			Role:     TenantMemberRoleMember,
			Status:   TenantMemberStatusActive,
		})
	}
	if err := DB.CreateInBatches(members, 100).Error; err != nil {
		return err
	}

	common.SysLog(fmt.Sprintf("migrated %d users to default tenant", len(userIds)))
	return nil
}
