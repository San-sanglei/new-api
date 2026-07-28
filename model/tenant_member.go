package model

import (
	"errors"

	"gorm.io/gorm"
)

// TenantMemberRole 租户内角色
// 与系统级 Role（0/1/10/100）独立，仅在租户范围内生效。
type TenantMemberRole int

const (
	TenantMemberRoleMember TenantMemberRole = 1 // 普通成员
	TenantMemberRoleAdmin  TenantMemberRole = 2 // 租户管理员
	TenantMemberRoleOwner  TenantMemberRole = 3 // 租户所有者
)

// TenantMemberStatus 成员状态
type TenantMemberStatus int

const (
	TenantMemberStatusActive   TenantMemberStatus = 1 // 活跃
	TenantMemberStatusInvited  TenantMemberStatus = 2 // 已邀请未接受
	TenantMemberStatusDisabled TenantMemberStatus = 3 // 已禁用
	TenantMemberStatusLeft     TenantMemberStatus = 4 // 已退出
)

// TenantMember 用户与租户的多对多关联表
// 一个用户可加入多个租户，User.CurrentTenantID 表示当前活动租户。
type TenantMember struct {
	Id        int                `json:"id" gorm:"primaryKey"`
	TenantId  int                `json:"tenant_id" gorm:"column:tenant_id;not null;uniqueIndex:idx_tenant_user"`
	UserId    int                `json:"user_id" gorm:"column:user_id;not null;uniqueIndex:idx_tenant_user;index"`
	Role      TenantMemberRole  `json:"role" gorm:"type:int;default:1"`
	Status    TenantMemberStatus `json:"status" gorm:"type:int;default:1"`
	JoinedAt  int64              `json:"joined_at" gorm:"autoCreateTime;column:joined_at"`
	CreatedAt int64              `json:"created_at" gorm:"autoCreateTime;column:created_at"`
	UpdatedAt int64              `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`
}

func (TenantMember) TableName() string {
	return "tenant_members"
}

// AddTenantMember 添加租户成员
func AddTenantMember(tenantId, userId int, role TenantMemberRole) error {
	member := TenantMember{
		TenantId: tenantId,
		UserId:   userId,
		Role:     role,
		Status:   TenantMemberStatusActive,
	}
	return DB.Create(&member).Error
}

// IsUserInTenant 检查用户是否属于指定租户（且状态为活跃）
func IsUserInTenant(userId, tenantId int) (bool, error) {
	if DB == nil {
		return false, errors.New("database not initialized")
	}
	var count int64
	err := DB.Model(&TenantMember{}).
		Where("user_id = ? AND tenant_id = ? AND status = ?", userId, tenantId, TenantMemberStatusActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserTenantIds 获取用户加入的所有租户 ID 列表
func GetUserTenantIds(userId int) ([]int, error) {
	var tenantIds []int
	err := DB.Model(&TenantMember{}).
		Where("user_id = ? AND status = ?", userId, TenantMemberStatusActive).
		Pluck("tenant_id", &tenantIds).Error
	return tenantIds, err
}

// GetUserTenantRole 查询用户在指定租户中的角色
func GetUserTenantRole(userId, tenantId int) (TenantMemberRole, error) {
	var member TenantMember
	err := DB.Where("user_id = ? AND tenant_id = ?", userId, tenantId).First(&member).Error
	if err != nil {
		return 0, err
	}
	return member.Role, nil
}

// GetUserCurrentTenantId 查询用户的当前活动租户 ID
// 优先使用 users.current_tenant_id；若该值为 0 或对应用户不在该租户中，回退到默认租户。
// 此函数供中间件调用，每个请求最多查询一次。
func GetUserCurrentTenantId(userId int) (int, error) {
	if DB == nil {
		return DefaultTenantID, errors.New("database not initialized")
	}
	if userId <= 0 {
		return DefaultTenantID, nil
	}

	var tenantId int
	err := DB.Model(&User{}).Where("id = ?", userId).Pluck("current_tenant_id", &tenantId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultTenantID, nil
		}
		return DefaultTenantID, err
	}
	if tenantId <= 0 {
		return DefaultTenantID, nil
	}

	// 校验用户仍在该租户中（活跃成员）
	inTenant, err := IsUserInTenant(userId, tenantId)
	if err != nil || !inTenant {
		return DefaultTenantID, nil
	}
	return tenantId, nil
}
