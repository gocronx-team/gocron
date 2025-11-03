package routers

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/host"
	"github.com/gocronx-team/gocron/internal/routers/install"
	"github.com/gocronx-team/gocron/internal/routers/loginlog"
	"github.com/gocronx-team/gocron/internal/routers/manage"
	"github.com/gocronx-team/gocron/internal/routers/task"
	"github.com/gocronx-team/gocron/internal/routers/tasklog"
	"github.com/gocronx-team/gocron/internal/routers/user"
	"github.com/rakyll/statik/fs"

	_ "github.com/gocronx-team/gocron/internal/statik"
)

const (
	urlPrefix = "/api"
	staticDir = "public"
)

var statikFS http.FileSystem

func init() {
	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}

// Register 路由泣册
func Register(r *gin.Engine) {
	// 一键安装脚本和下载（必须在 api 之前，避免被 NoRoute 拦截）
	r.GET("/install/agent", host.GetInstallScript)
	r.GET("/download/agent", host.DownloadAgent)

	api := r.Group(urlPrefix)

	// 系统安装
	installGroup := api.Group("/install")
	{
		installGroup.POST("/store", install.Store)
		installGroup.GET("/status", func(c *gin.Context) {
			jsonResp := utils.JsonResponse{}
			c.String(http.StatusOK, jsonResp.Success("", app.Installed))
		})
	}


	// 用户
	userGroup := api.Group("/user")
	{
		userGroup.GET("", user.Index)
		userGroup.GET("/:id", user.Detail)
		userGroup.POST("/store", user.Store)
		userGroup.POST("/remove/:id", user.Remove)
		userGroup.POST("/login", user.ValidateLogin)
		userGroup.POST("/enable/:id", user.Enable)
		userGroup.POST("/disable/:id", user.Disable)
		userGroup.POST("/editMyPassword", user.UpdateMyPassword)
		userGroup.POST("/editPassword/:id", user.UpdatePassword)
		// 2FA相关路由
		userGroup.GET("/2fa/status", user.Get2FAStatus)
		userGroup.GET("/2fa/setup", user.Setup2FA)
		userGroup.POST("/2fa/enable", user.Enable2FA)
		userGroup.POST("/2fa/disable", user.Disable2FA)
	}

	// 定时任务
	taskGroup := api.Group("/task")
	{
		taskGroup.POST("/store", task.Store)
		taskGroup.GET("/:id", task.Detail)
		taskGroup.GET("", task.Index)
		taskGroup.GET("/log", tasklog.Index)
		taskGroup.POST("/log/clear", tasklog.Clear)
		taskGroup.POST("/log/stop", tasklog.Stop)
		taskGroup.POST("/remove/:id", task.Remove)
		taskGroup.POST("/enable/:id", task.Enable)
		taskGroup.POST("/disable/:id", task.Disable)
		taskGroup.POST("/batch-enable", task.BatchEnable)
		taskGroup.POST("/batch-disable", task.BatchDisable)
		taskGroup.POST("/batch-remove", task.BatchRemove)
		taskGroup.GET("/run/:id", task.Run)
	}

	// 主机
	hostGroup := api.Group("/host")
	{
		hostGroup.GET("/:id", host.Detail)
		hostGroup.POST("/store", host.Store)
		hostGroup.GET("", host.Index)
		hostGroup.GET("/all", host.All)
		hostGroup.GET("/ping/:id", host.Ping)
		hostGroup.POST("/remove/:id", host.Remove)
		hostGroup.POST("/register", host.Register) // agent自动注册
		hostGroup.GET("/register-token", host.GetRegisterToken) // 获取注册 token
		hostGroup.POST("/register-token/generate", host.GenerateRegisterToken) // 生成注册 token
	}

	// 管理
	systemGroup := api.Group("/system")
	{
		slackGroup := systemGroup.Group("/slack")
		{
			slackGroup.GET("", manage.Slack)
			slackGroup.POST("/update", manage.UpdateSlack)
			slackGroup.POST("/channel", manage.CreateSlackChannel)
			slackGroup.POST("/channel/remove/:id", manage.RemoveSlackChannel)
		}
		mailGroup := systemGroup.Group("/mail")
		{
			mailGroup.GET("", manage.Mail)
			mailGroup.POST("/update", manage.UpdateMail)
			mailGroup.POST("/user", manage.CreateMailUser)
			mailGroup.POST("/user/remove/:id", manage.RemoveMailUser)
		}
		webhookGroup := systemGroup.Group("/webhook")
		{
			webhookGroup.GET("", manage.WebHook)
			webhookGroup.POST("/update", manage.UpdateWebHook)
		}
		systemGroup.GET("/login-log", loginlog.Index)
		systemGroup.GET("/log-retention", manage.GetLogRetentionDays)
		systemGroup.POST("/log-retention", manage.UpdateLogRetentionDays)
	}

	// API
	v1Group := api.Group("/v1")
	v1Group.Use(apiAuth)
	{
		v1Group.POST("/tasklog/remove/:id", tasklog.Remove)
		v1Group.POST("/task/enable/:id", task.Enable)
		v1Group.POST("/task/disable/:id", task.Disable)
	}

	// 首页路由（根路径）
	r.GET("/", func(c *gin.Context) {
		file, err := statikFS.Open("/index.html")
		if err != nil {
			logger.Error("读取首页文件失败: %s", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		c.Header("Content-Type", "text/html")
		io.Copy(c.Writer, file)
	})

	// 静态文件路由 - 必须放在最后
	r.NoRoute(func(c *gin.Context) {
		filepath := c.Request.URL.Path

		// 安装脚本路径已经在前面注册，不应该到这里
		// 如果到这里说明路由没有匹配，直接返回404
		if strings.HasPrefix(filepath, "/install/") {
			jsonResp := utils.JsonResponse{}
			c.String(http.StatusNotFound, jsonResp.Failure(utils.NotFound, "Install script not found"))
			return
		}

		// 移除 /public 前缀（如果存在）
		filepath = strings.TrimPrefix(filepath, "/public")

		// 尝试从 statikFS 读取文件
		file, err := statikFS.Open(filepath)
		if err == nil {
			defer file.Close()

			// 设置正确的Content-Type - 必须在写入数据之前设置
			if strings.HasSuffix(filepath, ".js") {
				c.Writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			} else if strings.HasSuffix(filepath, ".css") {
				c.Writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			} else if strings.HasSuffix(filepath, ".html") {
				c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			} else if strings.HasSuffix(filepath, ".png") {
				c.Writer.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(filepath, ".jpg") || strings.HasSuffix(filepath, ".jpeg") {
				c.Writer.Header().Set("Content-Type", "image/jpeg")
			} else if strings.HasSuffix(filepath, ".svg") {
				c.Writer.Header().Set("Content-Type", "image/svg+xml")
			}

			c.Status(http.StatusOK)
			io.Copy(c.Writer, file)
			return
		}

		// 文件不存在，返回404
		jsonResp := utils.JsonResponse{}
		c.String(http.StatusNotFound, jsonResp.Failure(utils.NotFound, i18n.T(c, "page_not_found")))
	})
}

// 中间件注册
func RegisterMiddleware(r *gin.Engine) {
	// 中间件
	r.Use(checkAppInstall)
	r.Use(ipAuth)
	r.Use(userAuth)
	r.Use(urlAuth)
}

// region 自定义中间件



/** 检测应用是否已安装 **/
func checkAppInstall(c *gin.Context) {
	if app.Installed {
		c.Next()
		return
	}
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/install") || path == "/" || strings.HasPrefix(path, "/static") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") {
		c.Next()
		return
	}
	jsonResp := utils.JsonResponse{}
	data := jsonResp.Failure(utils.AppNotInstall, i18n.T(c, "app_not_installed"))
	c.String(http.StatusOK, data)
	c.Abort()
}

// IP验证, 通过反向代理访问gocron，需设置Header X-Real-IP才能获取到客户端真实IP
func ipAuth(c *gin.Context) {
	if !app.Installed {
		c.Next()
		return
	}
	allowIpsStr := app.Setting.AllowIps
	if allowIpsStr == "" {
		c.Next()
		return
	}
	clientIp := c.ClientIP()
	allowIps := strings.Split(allowIpsStr, ",")
	if utils.InStringSlice(allowIps, clientIp) {
		c.Next()
		return
	}
	logger.Warnf("非法IP访问-%s", clientIp)
	jsonResp := utils.JsonResponse{}
	data := jsonResp.Failure(utils.UnauthorizedError, i18n.T(c, "unauthorized"))
	c.String(http.StatusOK, data)
	c.Abort()
}

// 用户认证
func userAuth(c *gin.Context) {
	if !app.Installed {
		c.Next()
		return
	}

	path := c.Request.URL.Path
	// 静态文件、安装脚本和下载直接放行
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".svg") || strings.HasPrefix(path, "/install/") || strings.HasPrefix(path, "/download/") {
		c.Next()
		return
	}

	uri := strings.TrimRight(path, "/")
	// 登录接口、安装状态接口、agent注册接口、安装脚本和下载不需要认证
	excludePaths := []string{"", "/api/user/login", "/api/install/status", "/api/host/register", "/install/agent", "/download/agent"}
	for _, p := range excludePaths {
		if uri == p {
			c.Next()
			return
		}
	}

	// v1 API接口使用单独的认证
	if strings.HasPrefix(uri, "/v1") {
		c.Next()
		return
	}

	// 尝试从token恢复用户信息
	err := user.RestoreToken(c)
	if err != nil {
		logger.Warnf("token解析失败: %v, path: %s", err, path)
		jsonResp := utils.JsonResponse{}
		data := jsonResp.Failure(utils.AuthError, i18n.T(c, "auth_failed"))
		c.String(http.StatusOK, data)
		c.Abort()
		return
	}

	if !user.IsLogin(c) {
		jsonResp := utils.JsonResponse{}
		data := jsonResp.Failure(utils.AuthError, i18n.T(c, "auth_failed"))
		c.String(http.StatusOK, data)
		c.Abort()
		return
	}

	c.Next()
}

// URL权限验证
func urlAuth(c *gin.Context) {
	if !app.Installed {
		c.Next()
		return
	}

	path := c.Request.URL.Path
	// 静态文件、安装脚本和下载直接放行
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".svg") || strings.HasPrefix(path, "/install/") || strings.HasPrefix(path, "/download/") {
		c.Next()
		return
	}

	if user.IsAdmin(c) {
		c.Next()
		return
	}
	uri := strings.TrimRight(path, "/")
	if strings.HasPrefix(uri, "/v1") {
		c.Next()
		return
	}
	// 普通用户允许访问的URL地址
	allowPaths := []string{
		"",
		"/api/install/status",
		"/api/task",
		"/api/task/log",
		"/api/host",
		"/api/host/all",
		"/api/host/register",
		"/api/user/login",
		"/api/user/editMyPassword",
		"/api/user/2fa/status",
		"/api/user/2fa/setup",
		"/api/user/2fa/enable",
		"/api/user/2fa/disable",
	}
	for _, p := range allowPaths {
		if p == uri {
			c.Next()
			return
		}
	}

	jsonResp := utils.JsonResponse{}
	data := jsonResp.Failure(utils.UnauthorizedError, i18n.T(c, "unauthorized"))
	c.String(http.StatusOK, data)
	c.Abort()
}

/** API接口签名验证 **/
func apiAuth(c *gin.Context) {
	if !app.Installed {
		c.Next()
		return
	}
	if !app.Setting.ApiSignEnable {
		c.Next()
		return
	}
	apiKey := strings.TrimSpace(app.Setting.ApiKey)
	apiSecret := strings.TrimSpace(app.Setting.ApiSecret)
	json := utils.JsonResponse{}
	if apiKey == "" || apiSecret == "" {
		msg := json.CommonFailure(i18n.T(c, "api_key_required"))
		c.String(http.StatusOK, msg)
		c.Abort()
		return
	}
	currentTimestamp := time.Now().Unix()
	timeParam, err := strconv.ParseInt(c.Query("time"), 10, 64)
	if err != nil || timeParam <= 0 {
		msg := json.CommonFailure(i18n.T(c, "param_time_required"))
		c.String(http.StatusOK, msg)
		c.Abort()
		return
	}
	if timeParam < (currentTimestamp - 1800) {
		msg := json.CommonFailure(i18n.T(c, "param_time_invalid"))
		c.String(http.StatusOK, msg)
		c.Abort()
		return
	}
	sign := strings.TrimSpace(c.Query("sign"))
	if sign == "" {
		msg := json.CommonFailure(i18n.T(c, "param_sign_required"))
		c.String(http.StatusOK, msg)
		c.Abort()
		return
	}
	raw := apiKey + strconv.FormatInt(timeParam, 10) + strings.TrimSpace(c.Request.URL.Path) + apiSecret
	realSign := utils.Md5(raw)
	if sign != realSign {
		msg := json.CommonFailure(i18n.T(c, "sign_verify_failed"))
		c.String(http.StatusOK, msg)
		c.Abort()
		return
	}
	c.Next()
}

// endregion
