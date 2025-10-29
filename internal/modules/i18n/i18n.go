package i18n

import "strings"

type Locale string

const (
	ZhCN Locale = "zh-CN"
	EnUS Locale = "en-US"
)

var messages = map[Locale]map[string]string{
	ZhCN: {
		"app_not_installed":        "应用未安装",
		"unauthorized":             "您无权限访问",
		"auth_failed":              "认证失败",
		"form_validation_failed":   "表单验证失败, 请检测输入",
		"username_exists":          "用户名已存在",
		"email_exists":             "邮箱已存在",
		"password_required":        "请输入密码",
		"password_confirm_required": "请再次输入密码",
		"password_mismatch":        "两次密码输入不一致",
		"old_password_error":       "原密码输入错误",
		"password_same_as_old":     "原密码与新密码不能相同",
		"username_password_empty":  "用户名、密码不能为空",
		"username_password_error":  "用户名或密码错误",
		"2fa_code_error":           "2FA验证码错误",
		"2fa_code_required":        "需要输入2FA验证码",
		"save_success":             "保存成功",
		"save_failed":              "保存失败",
		"update_success":           "修改成功",
		"update_failed":            "修改失败",
		"delete_success":           "删除成功",
		"delete_failed":            "删除失败",
		"operation_failed":         "操作失败",
		"api_key_required":         "使用API前, 请先配置密钥",
		"param_time_required":      "参数time不能为空",
		"param_time_invalid":       "time无效",
		"param_sign_required":      "参数sign不能为空",
		"sign_verify_failed":       "签名验证失败",
		"page_not_found":           "您访问的页面不存在",
	},
	EnUS: {
		"app_not_installed":        "Application not installed",
		"unauthorized":             "You do not have permission to access",
		"auth_failed":              "Authentication failed",
		"form_validation_failed":   "Form validation failed, please check your input",
		"username_exists":          "Username already exists",
		"email_exists":             "Email already exists",
		"password_required":        "Please enter password",
		"password_confirm_required": "Please enter password again",
		"password_mismatch":        "Passwords do not match",
		"old_password_error":       "Old password is incorrect",
		"password_same_as_old":     "New password cannot be the same as old password",
		"username_password_empty":  "Username and password cannot be empty",
		"username_password_error":  "Username or password is incorrect",
		"2fa_code_error":           "2FA verification code is incorrect",
		"2fa_code_required":        "2FA verification code is required",
		"save_success":             "Saved successfully",
		"save_failed":              "Save failed",
		"update_success":           "Updated successfully",
		"update_failed":            "Update failed",
		"delete_success":           "Deleted successfully",
		"delete_failed":            "Delete failed",
		"operation_failed":         "Operation failed",
		"api_key_required":         "Please configure API key before using API",
		"param_time_required":      "Parameter 'time' is required",
		"param_time_invalid":       "Parameter 'time' is invalid",
		"param_sign_required":      "Parameter 'sign' is required",
		"sign_verify_failed":       "Signature verification failed",
		"page_not_found":           "The page you are looking for does not exist",
	},
}

func T(locale Locale, key string) string {
	if msgs, ok := messages[locale]; ok {
		if msg, exists := msgs[key]; exists {
			return msg
		}
	}
	return key
}

func GetLocaleFromHeader(acceptLanguage string) Locale {
	if strings.Contains(strings.ToLower(acceptLanguage), "zh") {
		return ZhCN
	}
	return EnUS
}
