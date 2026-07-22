package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

// UpdateMidjourneyTaskBulk 定时轮询未完成的 Midjourney 任务。
// P1 修复：接受 ctx，服务退出时可通过 ctx.Done() 停止循环。
func UpdateMidjourneyTaskBulk(ctx context.Context) {
	//imageModel := "midjourney"
	for {
		if !common.SleepWithContext(ctx, time.Duration(15)*time.Second) {
			common.SysLog("UpdateMidjourneyTaskBulk stopping: context cancelled")
			return
		}

		// P4-3 修复：接收 error 以区分"无任务"与"DB 查询错误"
		tasks, err := model.GetAllUnFinishTasks()
		if err != nil {
			// DB error：记录日志，跳过本轮轮询（不影响下一轮）
			logger.LogError(ctx, fmt.Sprintf("UpdateMidjourneyTaskBulk: GetAllUnFinishTasks DB error err=%v", err))
			continue
		}
		if len(tasks) == 0 {
			continue
		}

		logger.LogInfo(ctx, fmt.Sprintf("检测到未完成的任务数有: %v", len(tasks)))
		taskChannelM := make(map[int][]string)
		taskM := make(map[string]*model.Midjourney)
		nullTaskIds := make([]int, 0)
		for _, task := range tasks {
			if task.MjId == "" {
				// 统计失败的未完成任务
				nullTaskIds = append(nullTaskIds, task.Id)
				continue
			}
			taskM[task.MjId] = task
			taskChannelM[task.ChannelId] = append(taskChannelM[task.ChannelId], task.MjId)
		}
		if len(nullTaskIds) > 0 {
			err := model.MjBulkUpdateByTaskIds(nullTaskIds, map[string]any{
				"status":   "FAILURE",
				"progress": "100%",
			})
			if err != nil {
				logger.LogError(ctx, fmt.Sprintf("Fix null mj_id task error: %v", err))
			} else {
				logger.LogInfo(ctx, fmt.Sprintf("Fix null mj_id task success: %v", nullTaskIds))
			}
		}
		if len(taskChannelM) == 0 {
			continue
		}

		for channelId, taskIds := range taskChannelM {
			logger.LogInfo(ctx, fmt.Sprintf("渠道 #%d 未完成的任务有: %d", channelId, len(taskIds)))
			if len(taskIds) == 0 {
				continue
			}
			midjourneyChannel, err := model.CacheGetChannel(channelId)
			if err != nil {
				logger.LogError(ctx, fmt.Sprintf("CacheGetChannel: %v", err))
				err := model.MjBulkUpdate(taskIds, map[string]any{
					"fail_reason": fmt.Sprintf("获取渠道信息失败，请联系管理员，渠道ID：%d", channelId),
					"status":      "FAILURE",
					"progress":    "100%",
				})
				if err != nil {
					logger.LogInfo(ctx, fmt.Sprintf("UpdateMidjourneyTask error: %v", err))
				}
				continue
			}
			requestUrl := fmt.Sprintf("%s/mj/task/list-by-condition", *midjourneyChannel.BaseURL)

			body, _ := json.Marshal(map[string]any{
				"ids": taskIds,
			})
			// P1-2 修复：使用 NewRequestWithContext 传递 ctx，后台循环退出时请求同步取消
			req, err := http.NewRequestWithContext(ctx, "POST", requestUrl, bytes.NewBuffer(body))
			if err != nil {
				logger.LogError(ctx, fmt.Sprintf("Get Task error: %v", err))
				continue
			}
			// 设置超时时间
			timeout := time.Second * 15
			reqCtx, cancel := context.WithTimeout(ctx, timeout)
			// 使用带有超时的 context 创建新的请求
			req = req.WithContext(reqCtx)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("mj-api-secret", midjourneyChannel.Key)
			resp, err := service.GetHttpClient().Do(req)
			if err != nil {
				logger.LogError(ctx, fmt.Sprintf("Get Task Do req error: %v", err))
				cancel()
				continue
			}
			if resp.StatusCode != http.StatusOK {
				logger.LogError(ctx, fmt.Sprintf("Get Task status code: %d", resp.StatusCode))
				resp.Body.Close()
				req.Body.Close()
				cancel()
				continue
			}
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.LogError(ctx, fmt.Sprintf("Get Mjp Task parse body error: %v", err))
				resp.Body.Close()
				req.Body.Close()
				cancel()
				continue
			}
			var responseItems []dto.MidjourneyDto
			err = json.Unmarshal(responseBody, &responseItems)
			if err != nil {
				logger.LogError(ctx, fmt.Sprintf("Get Mjp Task parse body error2: %v, body: %s", err, string(responseBody)))
				resp.Body.Close()
				req.Body.Close()
				cancel()
				continue
			}
			resp.Body.Close()
			req.Body.Close()
			cancel()

			for _, responseItem := range responseItems {
				task := taskM[responseItem.MjId]

				useTime := (time.Now().UnixNano() / int64(time.Millisecond)) - task.SubmitTime
				// 如果时间超过一小时，且进度不是100%，则认为任务失败
				if useTime > 3600000 && task.Progress != "100%" {
					responseItem.FailReason = "上游任务超时（超过1小时）"
					responseItem.Status = "FAILURE"
				}
				if !checkMjTaskNeedUpdate(task, responseItem) {
					continue
				}
				preStatus := task.Status
				task.Code = 1
				task.Progress = responseItem.Progress
				task.PromptEn = responseItem.PromptEn
				task.State = responseItem.State
				task.SubmitTime = responseItem.SubmitTime
				task.StartTime = responseItem.StartTime
				task.FinishTime = responseItem.FinishTime
				task.ImageUrl = responseItem.ImageUrl
				task.Status = responseItem.Status
				task.FailReason = responseItem.FailReason
				if responseItem.Properties != nil {
					propertiesStr, _ := json.Marshal(responseItem.Properties)
					task.Properties = string(propertiesStr)
				}
				if responseItem.Buttons != nil {
					buttonStr, _ := json.Marshal(responseItem.Buttons)
					task.Buttons = string(buttonStr)
				}
				// 映射 VideoUrl
				task.VideoUrl = responseItem.VideoUrl

				// 映射 VideoUrls - 将数组序列化为 JSON 字符串
				if responseItem.VideoUrls != nil && len(responseItem.VideoUrls) > 0 {
					videoUrlsStr, err := json.Marshal(responseItem.VideoUrls)
					if err != nil {
						logger.LogError(ctx, fmt.Sprintf("序列化 VideoUrls 失败: %v", err))
						task.VideoUrls = "[]" // 失败时设置为空数组
					} else {
						task.VideoUrls = string(videoUrlsStr)
					}
				} else {
					task.VideoUrls = "" // 空值时清空字段
				}

				shouldReturnQuota := false
				if (task.Progress != "100%" && responseItem.FailReason != "") || (task.Progress == "100%" && task.Status == "FAILURE") {
					logger.LogInfo(ctx, task.MjId+" 构建失败，"+task.FailReason)
					task.Progress = "100%"
					if task.Quota != 0 {
						shouldReturnQuota = true
					}
				}
				won, err := task.UpdateWithStatus(preStatus)
				if err != nil {
					logger.LogError(ctx, "UpdateMidjourneyTask task error: "+err.Error())
				} else if won && shouldReturnQuota {
					err = model.IncreaseUserQuota(task.UserId, task.Quota, false)
					if err != nil {
						// P4-5：退款失败但下方 RecordTaskBillingLog 仍会写入 LogTypeRefund 审计记录，
						//   造成"日志显示已退款但用户实际未收到退款"的审计断裂。
						//   升级为 SysError + billing_log_inconsistent=true 触发告警对账；
						//   不新增 return，不改变 UpdateMidjourneyTask 主流程。
						common.SysError(fmt.Sprintf(
							"Midjourney refund: user quota increase failed: refund_required=true billing_log_inconsistent=true user_id=%d task_id=%s quota=%d channel_id=%d err=%v",
							task.UserId, task.MjId, task.Quota, task.ChannelId, err,
						))
					}
					if err := model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
						UserId:    task.UserId,
						LogType:   model.LogTypeRefund,
						Content:   "",
						ChannelId: task.ChannelId,
						ModelName: service.CovertMjpActionToModelName(task.Action),
						Quota:     task.Quota,
						Other: map[string]interface{}{
							"task_id": task.MjId,
							"reason":  "构图失败",
						},
					}); err != nil {
						// P4-5: MJ 退款审计日志写入失败不阻塞退款流程，但必须 SysError 记录上下文供对账。
						common.SysError(fmt.Sprintf(
							"Midjourney refund: billing log persist failed: log_persist_required=true, task_id=%s, user_id=%d, log_type=%d, quota=%d, channel_id=%d, err=%v",
							task.MjId, task.UserId, model.LogTypeRefund, task.Quota, task.ChannelId, err,
						))
					}
				}
			}
		}
	}
}

func checkMjTaskNeedUpdate(oldTask *model.Midjourney, newTask dto.MidjourneyDto) bool {
	if oldTask.Code != 1 {
		return true
	}
	if oldTask.Progress != newTask.Progress {
		return true
	}
	if oldTask.PromptEn != newTask.PromptEn {
		return true
	}
	if oldTask.State != newTask.State {
		return true
	}
	if oldTask.SubmitTime != newTask.SubmitTime {
		return true
	}
	if oldTask.StartTime != newTask.StartTime {
		return true
	}
	if oldTask.FinishTime != newTask.FinishTime {
		return true
	}
	if oldTask.ImageUrl != newTask.ImageUrl {
		return true
	}
	if oldTask.Status != newTask.Status {
		return true
	}
	if oldTask.FailReason != newTask.FailReason {
		return true
	}
	if oldTask.FinishTime != newTask.FinishTime {
		return true
	}
	if oldTask.Progress != "100%" && newTask.FailReason != "" {
		return true
	}
	// 检查 VideoUrl 是否需要更新
	if oldTask.VideoUrl != newTask.VideoUrl {
		return true
	}
	// 检查 VideoUrls 是否需要更新
	if newTask.VideoUrls != nil && len(newTask.VideoUrls) > 0 {
		newVideoUrlsStr, _ := json.Marshal(newTask.VideoUrls)
		if oldTask.VideoUrls != string(newVideoUrlsStr) {
			return true
		}
	} else if oldTask.VideoUrls != "" {
		// 如果新数据没有 VideoUrls 但旧数据有，需要更新（清空）
		return true
	}

	return false
}

func GetAllMidjourney(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	// 解析其他查询参数
	queryParams := model.TaskQueryParams{
		ChannelID:      c.Query("channel_id"),
		MjID:           c.Query("mj_id"),
		StartTimestamp: c.Query("start_timestamp"),
		EndTimestamp:   c.Query("end_timestamp"),
	}

	// P4-3 修复：接收 error 以区分"无任务"与"DB 查询错误"
	items, err := model.GetAllTasks(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("GetAllMidjourney: GetAllTasks DB error err=%v", err))
		common.ApiErrorMsg(c, "获取 Midjourney 任务列表失败")
		return
	}
	total, err := model.CountAllTasks(queryParams)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("GetAllMidjourney: CountAllTasks DB error err=%v", err))
		common.ApiErrorMsg(c, "获取 Midjourney 任务总数失败")
		return
	}

	if setting.MjForwardUrlEnabled {
		for i, midjourney := range items {
			midjourney.ImageUrl = system_setting.ServerAddress + "/mj/image/" + midjourney.MjId
			items[i] = midjourney
		}
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func GetUserMidjourney(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	userId := c.GetInt("id")

	queryParams := model.TaskQueryParams{
		MjID:           c.Query("mj_id"),
		StartTimestamp: c.Query("start_timestamp"),
		EndTimestamp:   c.Query("end_timestamp"),
	}

	// P4-3 修复：接收 error 以区分"无任务"与"DB 查询错误"
	items, err := model.GetAllUserTask(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), queryParams)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("GetUserMidjourney: GetAllUserTask DB error userId=%d err=%v", userId, err))
		common.ApiErrorMsg(c, "获取用户 Midjourney 任务列表失败")
		return
	}
	total, err := model.CountAllUserTask(userId, queryParams)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("GetUserMidjourney: CountAllUserTask DB error userId=%d err=%v", userId, err))
		common.ApiErrorMsg(c, "获取用户 Midjourney 任务总数失败")
		return
	}

	if setting.MjForwardUrlEnabled {
		for i, midjourney := range items {
			midjourney.ImageUrl = system_setting.ServerAddress + "/mj/image/" + midjourney.MjId
			items[i] = midjourney
		}
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}
