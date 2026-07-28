package model

import (
	"errors"

	"gorm.io/gorm"
)

// TenantStatus 租户状态
type TenantStatus int

const (
	TenantStatusActive    TenantStatus = 1 // 活跃
	TenantStatusSuspended TenantStatus = 2 // 暂停
	TenantStatusDeleted   TenantStatus = 3 // 删除（软删除）
)

// Tenant 租户表
// 第一阶段：仅创建表与默认租户，不参与业务查询。
// 第二阶段起：业务表通过 tenant_id 与本表关联实现隔离。
type Tenant struct {
	Id          int           `json:"id" gorm:"primaryKey"`
	Name        string        `json:"name" gorm:"type:varchar(128);not null"` // 不加 unique，允许同名
	Slug        string        `json:"slug" gorm:"type:varchar(64);not null;uniqueIndex"`
	Status      TenantStatus  `json:"status" gorm:"type:int;default:1;index"`
	OwnerUserId int           `json:"owner_user_id" gorm:"column:owner_user_id;index"`
	CreatedAt   int64         `json:"created_at" gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   int64         `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Tenant) TableName() string {
	return "tenants"
}

// 默认租户常量
const (
	DefaultTenantID   = 1
	DefaultTenantName = "Took Official"
	DefaultTenantSlug = "took-official"
)

// EnsureDefaultTenant 幂等创建默认租户（id=1）
// 所有历史数据在迁移阶段写入 tenant_id=1，归属于本租户。
func EnsureDefaultTenant() error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	var count int64
	if err := DB.Model(&Tenant{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaultTenant := Tenant{
		Id:          DefaultTenantID,
		Name:        DefaultTenantName,
		Slug:        DefaultTenantSlug,
		Status:      TenantStatusActive,
		OwnerUserId: 1, // root user
	}
	return DB.Create(&defaultTenant).Error
}

// GetTenantById 查询租户
func GetTenantById(id int) (*Tenant, error) {
	var tenant Tenant
	if err := DB.First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetTenantBySlug 按 slug 查询租户
func GetTenantBySlug(slug string) (*Tenant, error) {
	var tenant Tenant
	if err := DB.Where("slug = ?", slug).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}
