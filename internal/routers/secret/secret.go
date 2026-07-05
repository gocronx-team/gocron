// Package secret 提供机密(Secret)的增删查接口。
//
// 机密值以 AES-GCM 密文存储,写入后不可读取(类 GitHub Secrets),仅在任务调度执行
// 时按需解密注入。本接口仅管理员可访问(普通用户不在 urlAuth 白名单内)。
package secret

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/crypto"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
)

// envNamePattern 限定机密名为合法的环境变量名(字母/下划线开头,后接字母数字下划线)。
var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// SecretForm 机密表单。
type SecretForm struct {
	Id     int    `form:"id" json:"id"`
	Name   string `form:"name" json:"name" binding:"required,max=64"`
	Value  string `form:"value" json:"value"`
	Remark string `form:"remark" json:"remark" binding:"max=255"`
}

// Index 机密列表(不含密文)。
func Index(c *gin.Context) {
	secretModel := new(models.Secret)
	list, err := secretModel.List()
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	base.RespondSuccess(c, utils.SuccessContent, map[string]interface{}{
		"total": len(list),
		"data":  list,
	})
}

// Store 创建或更新机密。
func Store(c *gin.Context) {
	var form SecretForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Remark = strings.TrimSpace(form.Remark)

	if !envNamePattern.MatchString(form.Name) {
		base.RespondError(c, i18n.T(c, "secret_name_invalid"))
		return
	}

	// 拒绝覆盖系统关键环境变量(PATH/LD_PRELOAD 等),防止破坏任务执行或库劫持
	if models.IsReservedEnvName(form.Name) {
		base.RespondError(c, i18n.T(c, "secret_name_reserved"))
		return
	}

	// 加密能力未配置时,无法安全存储,直接拒绝并提示。
	if !crypto.Configured() {
		base.RespondError(c, i18n.T(c, "secret_key_not_configured"))
		return
	}

	secretModel := new(models.Secret)
	nameCount, err := secretModel.NameExists(form.Name, form.Id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	if nameCount > 0 {
		base.RespondError(c, i18n.T(c, "secret_name_exists"))
		return
	}

	if form.Id == 0 {
		// 创建:必须提供值
		if form.Value == "" {
			base.RespondError(c, i18n.T(c, "secret_value_required"))
			return
		}
		ciphertext, err := crypto.Encrypt(form.Value)
		if err != nil {
			base.RespondError(c, i18n.T(c, "secret_encrypt_failed"), err)
			return
		}
		secretModel.Name = form.Name
		secretModel.Value = ciphertext
		secretModel.Remark = form.Remark
		newId, err := secretModel.Create()
		if err != nil {
			base.RespondError(c, i18n.T(c, "save_failed"), err)
			return
		}
		c.Set("audit_target_id", newId)
		c.Set("audit_target_name", form.Name)
		base.RespondSuccess(c, i18n.T(c, "save_success"), nil)
		return
	}

	// 更新:名称/备注总是更新;值为空表示保持原值不变(类 GitHub Secrets,值不可回显)。
	updateData := models.CommonMap{
		"name":   form.Name,
		"remark": form.Remark,
	}
	if form.Value != "" {
		ciphertext, err := crypto.Encrypt(form.Value)
		if err != nil {
			base.RespondError(c, i18n.T(c, "secret_encrypt_failed"), err)
			return
		}
		updateData["value"] = ciphertext
	}
	if _, err := secretModel.Update(form.Id, updateData); err != nil {
		base.RespondError(c, i18n.T(c, "update_failed"), err)
		return
	}
	c.Set("audit_target_id", form.Id)
	c.Set("audit_target_name", form.Name)
	base.RespondSuccess(c, i18n.T(c, "save_success"), nil)
}

// Remove 删除机密。
func Remove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	secretModel := new(models.Secret)
	if _, err := secretModel.Delete(id); err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	base.RespondSuccessWithDefaultMsg(c, nil)
}
