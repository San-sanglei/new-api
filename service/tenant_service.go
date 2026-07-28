package service

import (
	"github.com/gin-gonic/gin"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

// GetTenantID 从 gin.Context 获取当前请求的活动租户 ID
// 由 middleware/auth.go 在认证成功后写入。未命中时回退到默认租户（id=1）。
func GetTenantID(c *gin.Context) int {
	v, exists := c.Get(string(constant.ContextKeyTenantId))
	if !exists {
		return model.DefaultTenantID
	}
	tid, ok := v.(int)
	if !ok || tid <= 0 {
		return model.DefaultTenantID
	}
	return tid
}

// IsSuperAdmin 判断是否为超级管理员（可跨租户访问数据）
// 系统级 Role=100（RoleRootUser）才能跨租户。
func IsSuperAdmin(c *gin.Context) bool {
	return c.GetInt("role") >= common.RoleRootUser
}

// IsTenantMember 检查用户是否属于指定租户（活跃成员）
func IsTenantMember(userId, tenantId int) (bool, error) {
	return model.IsUserInTenant(userId, tenantId)
}

// CanAccessUserData 判断当前请求用户是否可以访问目标用户的数据
//
// 规则：
//   - 超级管理员（Role=100）：可访问任意用户
//   - 同一用户：可访问自己
//   - 其他：第一阶段不强制隔离，返回 true（保持向后兼容）
//
// 第二阶段实现租户内隔离后，将追加 tenant_id 校验逻辑。
func CanAccessUserData(c *gin.Context, targetUserId int) bool {
	if IsSuperAdmin(c) {
		return true
	}
	currentUserId := c.GetInt("id")
	if currentUserId == targetUserId {
		return true
	}
	// 第一阶段：不强制隔离，保持现有行为
	// 第二阶段：追加 tenant_id 校验
	return true
}
