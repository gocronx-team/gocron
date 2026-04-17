package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"gorm.io/gorm"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/routers/user"
	"github.com/gocronx-team/gocron/internal/service"
)

// VersionList 获取任务脚本版本列表
func VersionList(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if taskId <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	versionModel := new(models.TaskScriptVersion)
	params := models.CommonMap{}
	base.ParsePageAndPageSize(c, params)

	total, err := versionModel.Total(taskId)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	list, err := versionModel.List(taskId, params)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  list,
	})
	c.String(http.StatusOK, result)
}

// VersionDetail 获取单个版本详情
func VersionDetail(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	versionId, _ := strconv.Atoi(c.Param("version_id"))
	if taskId <= 0 || versionId <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	versionModel := new(models.TaskScriptVersion)
	version, err := versionModel.Detail(versionId)
	if err != nil || version.TaskId != taskId {
		base.RespondError(c, i18n.T(c, "version_not_found"))
		return
	}

	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, version)
	c.String(http.StatusOK, result)
}

// VersionRollback 回滚任务命令到指定版本
func VersionRollback(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	versionId, _ := strconv.Atoi(c.Param("version_id"))
	if taskId <= 0 || versionId <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	versionModel := new(models.TaskScriptVersion)
	version, err := versionModel.Detail(versionId)
	if err != nil || version.TaskId != taskId {
		base.RespondError(c, i18n.T(c, "version_not_found"))
		return
	}

	taskModel := new(models.Task)
	currentTask, err := taskModel.Detail(taskId)
	if err != nil || currentTask.Id == 0 {
		base.RespondError(c, i18n.T(c, "get_task_detail_failed"))
		return
	}

	// 使用事务保证回滚操作的原子性
	txErr := models.Db.Transaction(func(tx *gorm.DB) error {
		// 回滚前保存当前命令为新版本
		if currentTask.Command != version.Command {
			latestVersion, _ := versionModel.GetLatestVersion(taskId)
			saveVersion := &models.TaskScriptVersion{
				TaskId:   taskId,
				Command:  currentTask.Command,
				Remark:   "auto-save before rollback",
				Username: user.Username(c),
				Version:  latestVersion + 1,
			}
			if err := tx.Create(saveVersion).Error; err != nil {
				logger.Warnf("回滚前保存版本失败 TaskID-%d: %v", taskId, err)
			}
		}

		// 更新任务命令
		return tx.Model(&models.Task{}).Where("id = ?", taskId).
			UpdateColumn("command", version.Command).Error
	})
	if txErr != nil {
		base.RespondError(c, i18n.T(c, "rollback_failed"), txErr)
		return
	}

	// 事务完成后清理旧版本（非关键操作，不需要在事务内）
	if cErr := versionModel.CleanOldVersions(taskId, 30); cErr != nil {
		logger.Warnf("清理旧版本失败 TaskID-%d: %v", taskId, cErr)
	}

	// 重新加入调度器
	status, _ := taskModel.GetStatus(taskId)
	if status == models.Enabled {
		task, _ := taskModel.Detail(taskId)
		service.ServiceTask.RemoveAndAdd(task)
	}

	base.RespondSuccess(c, i18n.T(c, "rollback_success"), nil)
}
