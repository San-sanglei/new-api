package controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// UpdateTaskBulk 薄入口，实际轮询逻辑在 service 层。
// P1 修复：接受 ctx，传递给底层 TaskPollingLoop。
func UpdateTaskBulk(ctx context.Context) {
	service.TaskPollingLoop(ctx)
}

func GetAllTask(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	// 解析其他查询参数
	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ChannelID:      c.Query("channel_id"),
	}

	// Phase 2-D：管理员视角按租户过滤。Root 走跨租户查询。
	tenantId := 0
	if !service.IsSuperAdmin(c) {
		tenantId = service.GetTenantID(c)
	}

	items, err := model.TaskGetAllTasks(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams, tenantId)
	if err != nil {
		// P4-5 修复：DB error 时返回数据库错误响应，避免前端误以为"无任务"。
		common.SysError(fmt.Sprintf("GetAllTask: TaskGetAllTasks DB error: %v", err))
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	total := model.TaskCountAllTasks(queryParams, tenantId)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(tasksToDto(items, true))
	common.ApiSuccess(c, pageInfo)
}

func GetUserTask(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	userId := c.GetInt("id")

	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	queryParams := model.SyncTaskQueryParams{
		Platform:       constant.TaskPlatform(c.Query("platform")),
		TaskID:         c.Query("task_id"),
		Status:         c.Query("status"),
		Action:         c.Query("action"),
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}

	// Phase 2-D：用户视角按租户过滤。Root 走跨租户查询。
	tenantId := 0
	if !service.IsSuperAdmin(c) {
		tenantId = service.GetTenantID(c)
	}

	items, err := model.TaskGetAllUserTask(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams, tenantId)
	if err != nil {
		// P4-5 修复：DB error 时返回数据库错误响应，避免前端误以为"无任务"。
		common.SysError(fmt.Sprintf("GetUserTask: TaskGetAllUserTask DB error: user_id=%d, err=%v", userId, err))
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return
	}
	total := model.TaskCountAllUserTask(userId, queryParams, tenantId)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(tasksToDto(items, false))
	common.ApiSuccess(c, pageInfo)
}

func tasksToDto(tasks []*model.Task, fillUser bool) []*dto.TaskDto {
	var userIdMap map[int]*model.UserBase
	if fillUser {
		userIdMap = make(map[int]*model.UserBase)
		userIds := types.NewSet[int]()
		for _, task := range tasks {
			userIds.Add(task.UserId)
		}
		for _, userId := range userIds.Items() {
			cacheUser, err := model.GetUserCache(userId)
			if err == nil {
				userIdMap[userId] = cacheUser
			}
		}
	}
	result := make([]*dto.TaskDto, len(tasks))
	for i, task := range tasks {
		if fillUser {
			if user, ok := userIdMap[task.UserId]; ok {
				task.Username = user.Username
			}
		}
		result[i] = relay.TaskModel2Dto(task)
	}
	return result
}
