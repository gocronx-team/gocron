package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/pquerna/otp/totp"
)

const tokenDuration = 4 * time.Hour

// UserForm 用户表单
type UserForm struct {
	Id              int           `form:"id" json:"id"`
	Name            string        `form:"name" json:"name" binding:"required,max=32"` // 用户名
	Password        string        `form:"password" json:"password"` // 密码
	ConfirmPassword string        `form:"confirm_password" json:"confirm_password"` // 确认密码
	Email           string        `form:"email" json:"email" binding:"required,email,max=50"` // 邮箱
	IsAdmin         int8          `form:"is_admin" json:"is_admin"` // 是否是管理员 1:管理员 0:普通用户
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
		logger.Error(err)
	}
	total, err := userModel.Total()
	if err != nil {
		logger.Error(err)
	}

	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  users,
	})
	c.String(http.StatusOK, result)
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
	jsonResp := utils.JsonResponse{}
	var result string
	if userModel.Id == 0 {
		result = jsonResp.Success(utils.SuccessContent, nil)
	} else {
		result = jsonResp.Success(utils.SuccessContent, userModel)
	}
	c.String(http.StatusOK, result)
}

// 保存任务
func Store(c *gin.Context) {
	locale := getLocale(c)
	var form UserForm
	if err := c.ShouldBind(&form); err != nil {
		json := utils.JsonResponse{}
		result := json.CommonFailure(i18n.T(locale, "form_validation_failed"))
		c.String(http.StatusOK, result)
		return
	}
	
	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.TrimSpace(form.Email)
	form.Password = strings.TrimSpace(form.Password)
	form.ConfirmPassword = strings.TrimSpace(form.ConfirmPassword)
	json := utils.JsonResponse{}
	userModel := models.User{}
	nameExists, err := userModel.UsernameExists(form.Name, form.Id)
	if err != nil {
		result := json.CommonFailure(utils.FailureContent, err)
		c.String(http.StatusOK, result)
		return
	}
	if nameExists > 0 {
		result := json.CommonFailure(i18n.T(locale, "username_exists"))
		c.String(http.StatusOK, result)
		return
	}

	emailExists, err := userModel.EmailExists(form.Email, form.Id)
	if err != nil {
		result := json.CommonFailure(utils.FailureContent, err)
		c.String(http.StatusOK, result)
		return
	}
	if emailExists > 0 {
		result := json.CommonFailure(i18n.T(locale, "email_exists"))
		c.String(http.StatusOK, result)
		return
	}

	if form.Id == 0 {
		if form.Password == "" {
			result := json.CommonFailure(i18n.T(locale, "password_required"))
			c.String(http.StatusOK, result)
			return
		}
		if form.ConfirmPassword == "" {
			result := json.CommonFailure(i18n.T(locale, "password_confirm_required"))
			c.String(http.StatusOK, result)
			return
		}
		if form.Password != form.ConfirmPassword {
			result := json.CommonFailure(i18n.T(locale, "password_mismatch"))
			c.String(http.StatusOK, result)
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
			result := json.CommonFailure(i18n.T(locale, "save_failed"), err)
			c.String(http.StatusOK, result)
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
			result := json.CommonFailure(i18n.T(locale, "update_failed"), err)
			c.String(http.StatusOK, result)
			return
		}
	}

	result := json.Success(i18n.T(locale, "save_success"), nil)
	c.String(http.StatusOK, result)
}

// 删除用户
func Remove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	json := utils.JsonResponse{}

	userModel := new(models.User)
	_, err := userModel.Delete(id)
	var result string
	if err != nil {
		result = json.CommonFailure(utils.FailureContent, err)
	} else {
		result = json.Success(utils.SuccessContent, nil)
	}
	c.String(http.StatusOK, result)
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
	json := utils.JsonResponse{}
	userModel := new(models.User)
	_, err := userModel.Update(id, models.CommonMap{
		"status": status,
	})
	var result string
	if err != nil {
		result = json.CommonFailure(utils.FailureContent, err)
	} else {
		result = json.Success(utils.SuccessContent, nil)
	}
	c.String(http.StatusOK, result)
}

// UpdatePassword 更新密码
func UpdatePassword(c *gin.Context) {
	locale := getLocale(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var form UpdatePasswordForm
	if err := c.ShouldBind(&form); err != nil {
		json := utils.JsonResponse{}
		result := json.CommonFailure(i18n.T(locale, "form_validation_failed"))
		c.String(http.StatusOK, result)
		return
	}
	
	json := utils.JsonResponse{}
	var result string
	if form.NewPassword != form.ConfirmNewPassword {
		result = json.CommonFailure(i18n.T(locale, "password_mismatch"))
		c.String(http.StatusOK, result)
		return
	}
	userModel := new(models.User)
	_, err := userModel.UpdatePassword(id, form.NewPassword)
	if err != nil {
		result = json.CommonFailure(i18n.T(locale, "update_failed"))
	} else {
		result = json.Success(i18n.T(locale, "update_success"), nil)
	}
	c.String(http.StatusOK, result)
}

// UpdateMyPassword 更新我的密码
func UpdateMyPassword(c *gin.Context) {
	locale := getLocale(c)
	var form UpdateMyPasswordForm
	if err := c.ShouldBind(&form); err != nil {
		json := utils.JsonResponse{}
		result := json.CommonFailure(i18n.T(locale, "form_validation_failed"))
		c.String(http.StatusOK, result)
		return
	}
	
	json := utils.JsonResponse{}
	var result string
	if form.NewPassword != form.ConfirmNewPassword {
		result = json.CommonFailure(i18n.T(locale, "password_mismatch"))
		c.String(http.StatusOK, result)
		return
	}
	if form.OldPassword == form.NewPassword {
		result = json.CommonFailure(i18n.T(locale, "password_same_as_old"))
		c.String(http.StatusOK, result)
		return
	}
	userModel := new(models.User)
	if !userModel.Match(Username(c), form.OldPassword) {
		result = json.CommonFailure(i18n.T(locale, "old_password_error"))
		c.String(http.StatusOK, result)
		return
	}
	_, err := userModel.UpdatePassword(Uid(c), form.NewPassword)
	if err != nil {
		result = json.CommonFailure(i18n.T(locale, "update_failed"))
	} else {
		result = json.Success(i18n.T(locale, "update_success"), nil)
	}
	c.String(http.StatusOK, result)
}

// ValidateLogin 验证用户登录
func ValidateLogin(c *gin.Context) {
	locale := getLocale(c)
	username := strings.TrimSpace(c.PostForm("username"))
	password := strings.TrimSpace(c.PostForm("password"))
	twoFactorCode := strings.TrimSpace(c.PostForm("two_factor_code"))
	json := utils.JsonResponse{}
	var result string
	if username == "" || password == "" {
		result = json.CommonFailure(i18n.T(locale, "username_password_empty"))
		c.String(http.StatusOK, result)
		return
	}
	userModel := new(models.User)
	if !userModel.Match(username, password) {
		result = json.CommonFailure(i18n.T(locale, "username_password_error"))
		c.String(http.StatusOK, result)
		return
	}

	// 检查是否启用2FA
	if userModel.TwoFactorOn == 1 {
		if twoFactorCode == "" {
			result = json.Success(i18n.T(locale, "2fa_code_required"), map[string]interface{}{
				"require_2fa": true,
			})
			c.String(http.StatusOK, result)
			return
		}
		// 验证TOTP码
		valid := totp.Validate(twoFactorCode, userModel.TwoFactorKey)
		if !valid {
			result = json.CommonFailure(i18n.T(locale, "2fa_code_error"))
			c.String(http.StatusOK, result)
			return
		}
	}

	loginLogModel := new(models.LoginLog)
	loginLogModel.Username = userModel.Name
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	loginLogModel.Ip = ip
	_, err := loginLogModel.Create()
	if err != nil {
		logger.Error("记录用户登录日志失败", err)
	}

	token, err := generateToken(userModel)
	if err != nil {
		logger.Errorf("生成jwt失败: %s", err)
		result = json.Failure(utils.AuthError, i18n.T(locale, "auth_failed"))
		c.String(http.StatusOK, result)
		return
	}

	result = json.Success(utils.SuccessContent, map[string]interface{}{
		"token":    token,
		"uid":      userModel.Id,
		"username": userModel.Name,
		"is_admin": userModel.IsAdmin,
	})
	c.String(http.StatusOK, result)
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
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(tokenDuration).Unix()
	claims["uid"] = user.Id
	claims["iat"] = time.Now().Unix()
	claims["issuer"] = "gocron"
	claims["username"] = user.Name
	claims["is_admin"] = user.IsAdmin
	token.Claims = claims

	return token.SignedString([]byte(app.Setting.AuthSecret))
}

func getLocale(c *gin.Context) i18n.Locale {
	if locale, exists := c.Get("locale"); exists {
		if l, ok := locale.(i18n.Locale); ok {
			return l
		}
	}
	return i18n.ZhCN
}

// 还原jwt
func RestoreToken(c *gin.Context) error {
	authToken := c.GetHeader("Auth-Token")
	if authToken == "" {
		return nil
	}
	token, err := jwt.Parse(authToken, func(*jwt.Token) (interface{}, error) {
		return []byte(app.Setting.AuthSecret), nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}
	c.Set("uid", int(claims["uid"].(float64)))
	c.Set("username", claims["username"])
	c.Set("is_admin", int(claims["is_admin"].(float64)))

	return nil
}
