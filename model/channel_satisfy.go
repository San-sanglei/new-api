package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

// IsChannelEnabledForGroupModel 判断指定渠道是否对 (group, model) 启用。
// P4-4 修复：返回 (bool, error) 以区分"渠道未启用"与"DB 查询错误"。
// - 渠道启用：返回 (true, nil)
// - 渠道未启用：返回 (false, nil)
// - DB 查询失败：返回 (false, err)，调用方必须处理避免误判为 disabled
func IsChannelEnabledForGroupModel(group string, modelName string, channelID int) (bool, error) {
	if group == "" || modelName == "" || channelID <= 0 {
		return false, nil
	}
	if !common.MemoryCacheEnabled {
		return isChannelEnabledForGroupModelDB(group, modelName, channelID)
	}

	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()

	if group2model2channels == nil {
		return false, nil
	}

	if isChannelIDInList(group2model2channels[group][modelName], channelID) {
		return true, nil
	}
	normalized := ratio_setting.FormatMatchingModelName(modelName)
	if normalized != "" && normalized != modelName {
		return isChannelIDInList(group2model2channels[group][normalized], channelID), nil
	}
	return false, nil
}

// IsChannelEnabledForAnyGroupModel 判断指定渠道是否对任一 (groups, model) 启用。
// P4-4 修复：返回 (bool, error) 以区分"未启用"与"DB 查询错误"。
// - 任一 group 启用：返回 (true, nil)
// - 全部 group 未启用：返回 (false, nil)
// - 任一 group DB 查询失败：返回 (false, err)（保守起见，遇到错误立即返回）
func IsChannelEnabledForAnyGroupModel(groups []string, modelName string, channelID int) (bool, error) {
	if len(groups) == 0 {
		return false, nil
	}
	for _, g := range groups {
		enabled, err := IsChannelEnabledForGroupModel(g, modelName, channelID)
		if err != nil {
			return false, err
		}
		if enabled {
			return true, nil
		}
	}
	return false, nil
}

// isChannelEnabledForGroupModelDB 直接查询 DB 判断渠道是否启用。
// P4-4 修复：返回 (bool, error) 以区分"渠道未启用"与"DB 查询错误"。
// - 查询成功且 count > 0：返回 (true, nil)
// - 查询成功且 count == 0：返回 (false, nil)
// - DB 查询失败：返回 (false, err)，禁止用 false 误判为 disabled
func isChannelEnabledForGroupModelDB(group string, modelName string, channelID int) (bool, error) {
	var count int64
	err := DB.Model(&Ability{}).
		Where(commonGroupCol+" = ? and model = ? and channel_id = ? and enabled = ?", group, modelName, channelID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	normalized := ratio_setting.FormatMatchingModelName(modelName)
	if normalized == "" || normalized == modelName {
		return false, nil
	}
	count = 0
	err = DB.Model(&Ability{}).
		Where(commonGroupCol+" = ? and model = ? and channel_id = ? and enabled = ?", group, normalized, channelID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func isChannelIDInList(list []int, channelID int) bool {
	for _, id := range list {
		if id == channelID {
			return true
		}
	}
	return false
}
