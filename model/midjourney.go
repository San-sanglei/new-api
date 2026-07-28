package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

type Midjourney struct {
	Id          int    `json:"id"`
	Code        int    `json:"code"`
	UserId      int    `json:"user_id" gorm:"index"`
	Action      string `json:"action" gorm:"type:varchar(40);index"`
	MjId        string `json:"mj_id" gorm:"index"`
	Prompt      string `json:"prompt"`
	PromptEn    string `json:"prompt_en"`
	Description string `json:"description"`
	State       string `json:"state"`
	SubmitTime  int64  `json:"submit_time" gorm:"index"`
	StartTime   int64  `json:"start_time" gorm:"index"`
	FinishTime  int64  `json:"finish_time" gorm:"index"`
	ImageUrl    string `json:"image_url"`
	VideoUrl    string `json:"video_url"`
	VideoUrls   string `json:"video_urls"`
	Status      string `json:"status" gorm:"type:varchar(20);index"`
	Progress    string `json:"progress" gorm:"type:varchar(30);index"`
	FailReason  string `json:"fail_reason"`
	ChannelId   int    `json:"channel_id"`
	Quota       int    `json:"quota"`
	Buttons     string `json:"buttons"`
	Properties  string `json:"properties"`

	// Phase 2-D：任务归属租户（默认租户=1，向后兼容）
	// 写入：mjproxy_handler 创建任务时从 RelayInfo.TenantID 获取
	// 查询：admin/用户列表按 tenant_id 过滤；轮询路径不过滤
	TenantID int `json:"tenant_id" gorm:"type:int;default:1;index"`
}

// TaskQueryParams 用于包含所有搜索条件的结构体，可以根据需求添加更多字段
type TaskQueryParams struct {
	ChannelID      string
	MjID           string
	StartTimestamp string
	EndTimestamp   string
}

// GetAllUserTask 查询用户 Midjourney 任务列表。
// P4-3 修复：返回 ([]*Midjourney, error) 以区分"无任务"与"DB 查询错误"。
// - 查询成功（含空结果）：返回 (tasks, nil)
// - DB error：返回 (nil, error)
//
// Phase 2-D：tenantId > 0 时按租户过滤；<= 0 时跨租户查询（Root）。
func GetAllUserTask(userId int, startIdx int, num int, queryParams TaskQueryParams, tenantId int) ([]*Midjourney, error) {
	var tasks []*Midjourney

	// 初始化查询构建器
	query := DB.Where("user_id = ?", userId)
	if tenantId > 0 {
		query = query.Where("tenant_id = ?", tenantId)
	}

	if queryParams.MjID != "" {
		query = query.Where("mj_id = ?", queryParams.MjID)
	}
	if queryParams.StartTimestamp != "" {
		// 假设您已将前端传来的时间戳转换为数据库所需的时间格式，并处理了时间戳的验证和解析
		query = query.Where("submit_time >= ?", queryParams.StartTimestamp)
	}
	if queryParams.EndTimestamp != "" {
		query = query.Where("submit_time <= ?", queryParams.EndTimestamp)
	}

	// 获取数据
	err := query.Order("id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetAllTasks 查询所有 Midjourney 任务列表（管理员视角）。
// P4-3 修复：返回 ([]*Midjourney, error) 以区分"无任务"与"DB 查询错误"。
// - 查询成功（含空结果）：返回 (tasks, nil)
// - DB error：返回 (nil, error)
//
// Phase 2-D：tenantId > 0 时按租户过滤；<= 0 时跨租户查询（Root）。
func GetAllTasks(startIdx int, num int, queryParams TaskQueryParams, tenantId int) ([]*Midjourney, error) {
	var tasks []*Midjourney

	// 初始化查询构建器
	query := DB
	if tenantId > 0 {
		query = query.Where("tenant_id = ?", tenantId)
	}

	// 添加过滤条件
	if queryParams.ChannelID != "" {
		query = query.Where("channel_id = ?", queryParams.ChannelID)
	}
	if queryParams.MjID != "" {
		query = query.Where("mj_id = ?", queryParams.MjID)
	}
	if queryParams.StartTimestamp != "" {
		query = query.Where("submit_time >= ?", queryParams.StartTimestamp)
	}
	if queryParams.EndTimestamp != "" {
		query = query.Where("submit_time <= ?", queryParams.EndTimestamp)
	}

	// 获取数据
	err := query.Order("id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetAllUnFinishTasks 查询所有未完成的 Midjourney 任务。
// P4-3 修复：返回 ([]*Midjourney, error) 以区分"无任务"与"DB 查询错误"。
// - 查询成功（含空结果）：返回 (tasks, nil)
// - DB error：返回 (nil, error)，调用方必须处理避免轮询循环静默跳过
func GetAllUnFinishTasks() ([]*Midjourney, error) {
	var tasks []*Midjourney
	// get all tasks progress is not 100%
	err := DB.Where("progress != ?", "100%").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByOnlyMJId 按 mj_id 查询 Midjourney 任务（不限定用户）。
// P4-3 修复：返回 (*Midjourney, error) 以区分"任务不存在"与"DB 查询错误"。
// - 任务不存在（ErrRecordNotFound）：返回 (nil, nil)
// - DB error：返回 (nil, error)
func GetByOnlyMJId(mjId string) (*Midjourney, error) {
	var mj *Midjourney
	err := DB.Where("mj_id = ?", mjId).First(&mj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mj, nil
}

// GetByMJId 按 user_id + mj_id 查询 Midjourney 任务。
// P4-3 修复：返回 (*Midjourney, error) 以区分"任务不存在"与"DB 查询错误"。
// - 任务不存在（ErrRecordNotFound）：返回 (nil, nil)
// - DB error：返回 (nil, error)
func GetByMJId(userId int, mjId string) (*Midjourney, error) {
	var mj *Midjourney
	err := DB.Where("user_id = ? and mj_id = ?", userId, mjId).First(&mj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mj, nil
}

// GetByMJIds 按 user_id + mj_ids 批量查询 Midjourney 任务。
// P4-3 修复：返回 ([]*Midjourney, error) 以区分"无任务"与"DB 查询错误"。
// - 查询成功（含空结果）：返回 (mj, nil)
// - DB error：返回 (nil, error)
func GetByMJIds(userId int, mjIds []string) ([]*Midjourney, error) {
	var mj []*Midjourney
	err := DB.Where("user_id = ? and mj_id in (?)", userId, mjIds).Find(&mj).Error
	if err != nil {
		return nil, err
	}
	return mj, nil
}

func GetMjByuId(id int) *Midjourney {
	var mj *Midjourney
	var err error
	err = DB.Where("id = ?", id).First(&mj).Error
	if err != nil {
		return nil
	}
	return mj
}

func UpdateProgress(id int, progress string) error {
	return DB.Model(&Midjourney{}).Where("id = ?", id).Update("progress", progress).Error
}

func (midjourney *Midjourney) Insert() error {
	var err error
	err = DB.Create(midjourney).Error
	return err
}

func (midjourney *Midjourney) Update() error {
	var err error
	err = DB.Save(midjourney).Error
	return err
}

// UpdateWithStatus performs a conditional UPDATE guarded by fromStatus (CAS).
// Returns (true, nil) if this caller won the update, (false, nil) if
// another process already moved the task out of fromStatus.
// UpdateWithStatus performs a conditional UPDATE guarded by fromStatus (CAS).
// Uses Model().Select("*").Updates() to avoid GORM Save()'s INSERT fallback.
func (midjourney *Midjourney) UpdateWithStatus(fromStatus string) (bool, error) {
	result := DB.Model(midjourney).Where("status = ?", fromStatus).Select("*").Updates(midjourney)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func MjBulkUpdate(mjIds []string, params map[string]any) error {
	return DB.Model(&Midjourney{}).
		Where("mj_id in (?)", mjIds).
		Updates(params).Error
}

func MjBulkUpdateByTaskIds(taskIDs []int, params map[string]any) error {
	return DB.Model(&Midjourney{}).
		Where("id in (?)", taskIDs).
		Updates(params).Error
}

// CountAllTasks 统计所有 Midjourney 任务总数（管理员视角）。
// P4-3 修复：返回 (int64, error) 以区分"0 任务"与"DB 查询错误"。
// - 统计成功：返回 (total, nil)
// - DB error：返回 (0, error)，避免调用方用 0 当作真实数据
//
// Phase 2-D：tenantId > 0 时按租户过滤；<= 0 时跨租户查询（Root）
func CountAllTasks(queryParams TaskQueryParams, tenantId int) (int64, error) {
	var total int64
	query := DB.Model(&Midjourney{})
	if tenantId > 0 {
		query = query.Where("tenant_id = ?", tenantId)
	}
	if queryParams.ChannelID != "" {
		query = query.Where("channel_id = ?", queryParams.ChannelID)
	}
	if queryParams.MjID != "" {
		query = query.Where("mj_id = ?", queryParams.MjID)
	}
	if queryParams.StartTimestamp != "" {
		query = query.Where("submit_time >= ?", queryParams.StartTimestamp)
	}
	if queryParams.EndTimestamp != "" {
		query = query.Where("submit_time <= ?", queryParams.EndTimestamp)
	}
	if err := query.Count(&total).Error; err != nil {
		common.SysError(fmt.Sprintf("CountAllTasks query failed: error=%s", err.Error()))
		return 0, err
	}
	return total, nil
}

// CountAllUserTask 统计用户 Midjourney 任务总数。
// P4-3 修复：返回 (int64, error) 以区分"0 任务"与"DB 查询错误"。
// - 统计成功：返回 (total, nil)
// - DB error：返回 (0, error)，避免调用方用 0 当作真实数据
//
// Phase 2-D：tenantId > 0 时按租户过滤；<= 0 时跨租户查询（Root）
func CountAllUserTask(userId int, queryParams TaskQueryParams, tenantId int) (int64, error) {
	var total int64
	query := DB.Model(&Midjourney{}).Where("user_id = ?", userId)
	if tenantId > 0 {
		query = query.Where("tenant_id = ?", tenantId)
	}
	if queryParams.MjID != "" {
		query = query.Where("mj_id = ?", queryParams.MjID)
	}
	if queryParams.StartTimestamp != "" {
		query = query.Where("submit_time >= ?", queryParams.StartTimestamp)
	}
	if queryParams.EndTimestamp != "" {
		query = query.Where("submit_time <= ?", queryParams.EndTimestamp)
	}
	if err := query.Count(&total).Error; err != nil {
		common.SysError(fmt.Sprintf("CountAllUserTask query failed: user_id=%d, error=%s", userId, err.Error()))
		return 0, err
	}
	return total, nil
}
