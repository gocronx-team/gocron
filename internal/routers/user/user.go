package user

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
)

const tokenDuration = 4 * time.Hour

// UserForm 用户表单
type UserForm struct {
	Id              int           `form:"id" json:"id"`
	Name            string        `form:"name" json:"name" binding:"required,max=32"`         // 用户名
	Password        string        `form:"password" json:"password"`                           // 密码
	ConfirmPassword string        `form:"confirm_password" json:"confirm_password"`           // 确认密码
	Email           string        `form:"email" json:"email" binding:"required,email,max=50"` // 邮箱
	IsAdmin         int8          `form:"is_admin" json:"is_admin"`                           // 是否是管理员 1:管理员 0:普通用户
	Status          models.Status `form:"status" json:"status"`
}

// UpdatePasswordForm 更新密码表单
type UpdatePasswordForm struct {
	NewPassword        string `form:"new_password" json:"new_password" binding:"required,min=6"`
	ConfirmNewPassword string `form:"confirm_new_password" json:"confirm_new_password" binding:"required,min=6"`
}

// UpdateMyPasswordForm 更新我的密码表单
type UpdateMyPasswordForm struct {
	OldPassword        string `form:"old_password" json:"old_password" binding:"required"`
	NewPassword        string `form:"new_password" json:"new_password" binding:"required,min=6"`
	ConfirmNewPassword string `form:"confirm_new_password" json:"confirm_new_password" binding:"required,min=6"`
}

// Index 用户列表页
func Index(c *gin.Context) {
	queryParams := parseQueryParams(c)
	userModel := new(models.User)
	users, err := userModel.List(queryParams)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	total, err := userModel.Total()
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	base.RespondSuccess(c, utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  users,
	})
}

// 解析查询参数
func parseQueryParams(c *gin.Context) models.CommonMap {
	params := models.CommonMap{}
	base.ParsePageAndPageSize(c, params)

	return params
}

// Detail 用户详情
func Detail(c *gin.Context) {
	userModel := new(models.User)
	id, _ := strconv.Atoi(c.Param("id"))
	err := userModel.Find(id)
	if err != nil {
		logger.Error(err)
	}
	if userModel.Id == 0 {
		base.RespondSuccess(c, utils.SuccessContent, nil)
	} else {
		base.RespondSuccess(c, utils.SuccessContent, userModel)
	}
}

// 保存任务
func Store(c *gin.Context) {
	var form UserForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.TrimSpace(form.Email)
	form.Password = strings.TrimSpace(form.Password)
	form.ConfirmPassword = strings.TrimSpace(form.ConfirmPassword)

	userModel := models.User{}
	nameExists, err := userModel.UsernameExists(form.Name, form.Id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	if nameExists > 0 {
		base.RespondError(c, i18n.T(c, "username_exists"))
		return
	}

	emailExists, err := userModel.EmailExists(form.Email, form.Id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	if emailExists > 0 {
		base.RespondError(c, i18n.T(c, "email_exists"))
		return
	}

	if form.Id == 0 {
		if form.Password == "" {
			base.RespondError(c, i18n.T(c, "password_required"))
			return
		}
		if form.ConfirmPassword == "" {
			base.RespondError(c, i18n.T(c, "password_confirm_required"))
			return
		}
		// 验证密码复杂度
		if valid, errKey := utils.ValidatePassword(form.Password); !valid {
			base.RespondError(c, i18n.T(c, errKey))
			return
		}
		if form.Password != form.ConfirmPassword {
			base.RespondError(c, i18n.T(c, "password_mismatch"))
			return
		}
	}
	userModel.Name = form.Name
	userModel.Email = form.Email
	userModel.Password = form.Password
	userModel.IsAdmin = form.IsAdmin
	userModel.Status = form.Status

	if form.Id == 0 {
		_, err = userModel.Create()
		if err != nil {
			base.RespondError(c, i18n.T(c, "save_failed"), err)
			return
		}
	} else {
		_, err = userModel.Update(form.Id, models.CommonMap{
			"name":     form.Name,
			"email":    form.Email,
			"status":   form.Status,
			"is_admin": form.IsAdmin,
		})
		if err != nil {
			base.RespondError(c, i18n.T(c, "update_failed"), err)
			return
		}
	}

	base.RespondSuccess(c, i18n.T(c, "save_success"), nil)
}

// 删除用户
func Remove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userModel := new(models.User)
	_, err := userModel.Delete(id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
	} else {
		base.RespondSuccessWithDefaultMsg(c, nil)
	}
}

// 激活用户
func Enable(c *gin.Context) {
	changeStatus(c, models.Enabled)
}

// 禁用用户
func Disable(c *gin.Context) {
	changeStatus(c, models.Disabled)
}

// 改变任务状态
func changeStatus(c *gin.Context, status models.Status) {
	id, _ := strconv.Atoi(c.Param("id"))
	userModel := new(models.User)
	_, err := userModel.Update(id, models.CommonMap{
		"status": status,
	})
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
	} else {
		base.RespondSuccessWithDefaultMsg(c, nil)
	}
}

// UpdatePassword 更新密码
func UpdatePassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var form UpdatePasswordForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	if form.NewPassword != form.ConfirmNewPassword {
		base.RespondError(c, i18n.T(c, "password_mismatch"))
		return
	}
	// 验证密码复杂度
	if valid, errKey := utils.ValidatePassword(form.NewPassword); !valid {
		base.RespondError(c, i18n.T(c, errKey))
		return
	}
	userModel := new(models.User)
	_, err := userModel.UpdatePassword(id, form.NewPassword)
	if err != nil {
		base.RespondError(c, i18n.T(c, "update_failed"))
	} else {
		base.RespondSuccess(c, i18n.T(c, "update_success"), nil)
	}
}

// UpdateMyPassword 更新我的密码
func UpdateMyPassword(c *gin.Context) {
	var form UpdateMyPasswordForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	if form.NewPassword != form.ConfirmNewPassword {
		base.RespondError(c, i18n.T(c, "password_mismatch"))
		return
	}
	if form.OldPassword == form.NewPassword {
		base.RespondError(c, i18n.T(c, "password_same_as_old"))
		return
	}
	// 验证密码复杂度
	if valid, errKey := utils.ValidatePassword(form.NewPassword); !valid {
		base.RespondError(c, i18n.T(c, errKey))
		return
	}
	userModel := new(models.User)
	if !userModel.Match(Username(c), form.OldPassword) {
		base.RespondError(c, i18n.T(c, "old_password_error"))
		return
	}
	_, err := userModel.UpdatePassword(Uid(c), form.NewPassword)
	if err != nil {
		base.RespondError(c, i18n.T(c, "update_failed"))
	} else {
		base.RespondSuccess(c, i18n.T(c, "update_success"), nil)
	}
}

// ValidateLogin 验证用户登录
func ValidateLogin(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := strings.TrimSpace(c.PostForm("password"))
	twoFactorCode := strings.TrimSpace(c.PostForm("two_factor_code"))

	if username == "" || password == "" {
		base.RespondError(c, i18n.T(c, "username_password_empty"))
		return
	}

	// 获取登录限制器
	limiter := utils.GetLoginLimiter()

	// 检查账户是否被锁定
	if locked, lockTime := limiter.IsLocked(username); locked {
		remainingTime := int(time.Until(lockTime).Minutes())
		if remainingTime < 1 {
			remainingTime = 1
		}
		base.RespondError(c, fmt.Sprintf(i18n.T(c, "account_locked"), remainingTime))
		return
	}

	userModel := new(models.User)
	if !userModel.Match(username, password) {
		// 记录登录失败
		limiter.RecordFailure(username)
		remaining := limiter.GetRemainingAttempts(username)

		if remaining > 0 {
			base.RespondError(c, fmt.Sprintf(i18n.T(c, "login_failed_with_attempts"), remaining))
		} else {
			base.RespondError(c, i18n.T(c, "username_password_error"))
		}
		return
	}

	// 检查是否启用2FA
	if userModel.TwoFactorOn == 1 {
		if twoFactorCode == "" {
			base.RespondSuccess(c, i18n.T(c, "2fa_code_required"), map[string]interface{}{
				"require_2fa": true,
			})
			return
		}
		// 验证TOTP码
		valid := totp.Validate(twoFactorCode, userModel.TwoFactorKey)
		if !valid {
			// 2FA验证失败也记录失败次数
			limiter.RecordFailure(username)
			base.RespondError(c, i18n.T(c, "2fa_code_error"))
			return
		}
	}

	// 登录成功，清除失败记录
	limiter.RecordSuccess(username)

	loginLogModel := new(models.LoginLog)
	loginLogModel.Username = userModel.Name
	loginLogModel.Ip = utils.ClientIP(c)
	_, err := loginLogModel.Create()
	if err != nil {
		logger.Error("记录用户登录日志失败", err)
	}

	token, err := generateToken(userModel)
	if err != nil {
		logger.Errorf("生成jwt失败: %s", err)
		base.RespondAuthError(c, i18n.T(c, "auth_failed"))
		return
	}

	base.RespondSuccess(c, utils.SuccessContent, map[string]interface{}{
		"token":    token,
		"uid":      userModel.Id,
		"username": userModel.Name,
		"is_admin": userModel.IsAdmin,
	})
}

// Username 获取session中的用户名
func Username(c *gin.Context) string {
	usernameInterface, ok := c.Get("username")
	if !ok {
		return ""
	}
	if username, ok := usernameInterface.(string); ok {
		return username
	} else {
		return ""
	}
}

// Uid 获取session中的Uid
func Uid(c *gin.Context) int {
	uidInterface, ok := c.Get("uid")
	if !ok {
		return 0
	}
	if uid, ok := uidInterface.(int); ok {
		return uid
	} else {
		return 0
	}
}

// IsLogin 判断用户是否已登录
func IsLogin(c *gin.Context) bool {
	return Uid(c) > 0
}

// IsAdmin 判断当前用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	isAdmin, ok := c.Get("is_admin")
	if !ok {
		return false
	}
	if v, ok := isAdmin.(int); ok {
		return v > 0
	} else {
		return false
	}
}

// 生成jwt
func generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"exp":      time.Now().Add(tokenDuration).Unix(),
		"uid":      user.Id,
		"iat":      time.Now().Unix(),
		"issuer":   "gocron",
		"username": user.Name,
		"is_admin": user.IsAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(app.Setting.AuthSecret))
}

// 还原jwt，如果 token 即将过期（小于1小时），则自动刷新
func RestoreToken(c *gin.Context) (string, error) {
	authToken := c.GetHeader("Auth-Token")
	if authToken == "" {
		return "", nil
	}
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(app.Setting.AuthSecret), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	uidF, ok := claims["uid"].(float64)
	if !ok {
		return "", errors.New("invalid uid claim")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("invalid username claim")
	}
	isAdminF, ok := claims["is_admin"].(float64)
	if !ok {
		return "", errors.New("invalid is_admin claim")
	}
	expF, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("invalid exp claim")
	}

	c.Set("uid", int(uidF))
	c.Set("username", username)
	c.Set("is_admin", int(isAdminF))

	// 检查 token 是否即将过期（小于 1 小时）
	exp := int64(expF)
	if time.Until(time.Unix(exp, 0)) < time.Hour {
		// 生成新 token
		userModel := &models.User{
			Id:      int(uidF),
			Name:    username,
			IsAdmin: int8(isAdminF),
		}
		newToken, err := generateToken(userModel)
		if err != nil {
			logger.Warnf("刷新token失败: %v", err)
			return "", nil
		}
		logger.Infof("用户 %s 的 token 已自动刷新", userModel.Name)
		return newToken, nil
	}

	return "", nil
}
