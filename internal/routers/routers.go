package routers

import (
	"context"
	"crypto/subtle"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gocronembed "github.com/gocronx-team/gocron"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/agent"
	"github.com/gocronx-team/gocron/internal/routers/audit"
	"github.com/gocronx-team/gocron/internal/routers/host"
	"github.com/gocronx-team/gocron/internal/routers/install"
	"github.com/gocronx-team/gocron/internal/routers/loginlog"
	"github.com/gocronx-team/gocron/internal/routers/manage"
	"github.com/gocronx-team/gocron/internal/routers/statistics"
	"github.com/gocronx-team/gocron/internal/routers/task"
	"github.com/gocronx-team/gocron/internal/routers/tasklog"
	"github.com/gocronx-team/gocron/internal/routers/template"
	"github.com/gocronx-team/gocron/internal/routers/user"
)

const (
	urlPrefix = "/api"
)

var staticFS fs.FS

func init() {
	var err error
	staticFS, err = gocronembed.StaticFS()
	if err != nil {
		logger.Fatal("初始化静态文件系统失败", err)
	}
}

// Register 路由泣册
func Register(r *gin.Engine) {
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
		taskGroup.GET("/versions/:id", task.VersionList)
		taskGroup.GET("/versions/:id/:version_id", task.VersionDetail)
		taskGroup.POST("/versions/:id/:version_id/rollback", task.VersionRollback)
		taskGroup.POST("/store", task.Store)
		taskGroup.POST("/cron-preview", task.CronPreview)
		taskGroup.GET("/tags", task.GetAllTags)
		taskGroup.GET("/:id", task.Detail)
		taskGroup.GET("", task.Index)
		taskGroup.GET("/log", tasklog.Index)
		taskGroup.POST("/log/clear", tasklog.Clear)
		taskGroup.POST("/log/clear/:id", tasklog.ClearByTaskId)
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
	}

	// Agent注册
	agentGroup := api.Group("/agent")
	{
		agentGroup.POST("/generate-token", agent.GenerateToken)
		agentGroup.GET("/install.sh", agent.InstallScript)
		agentGroup.POST("/register", agent.Register)
		agentGroup.GET("/download", agent.Download)
	}

	// 任务模板
	templateGroup := api.Group("/template")
	{
		templateGroup.GET("", template.Index)
		templateGroup.GET("/categories", template.Categories)
		templateGroup.GET("/:id", template.Detail)
		templateGroup.POST("/store", template.Store)
		templateGroup.POST("/remove/:id", template.Remove)
		templateGroup.POST("/apply/:id", template.Apply)
		templateGroup.POST("/save-from-task", template.SaveFromTask)
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
			webhookGroup.POST("/url", manage.CreateWebhookUrl)
			webhookGroup.POST("/url/remove/:id", manage.RemoveWebhookUrl)
		}
		systemGroup.GET("/login-log", loginlog.Index)
		systemGroup.GET("/log-retention", manage.GetLogRetentionDays)
		systemGroup.POST("/log-retention", manage.UpdateLogRetentionDays)
	}

	// 统计
	statisticsGroup := api.Group("/statistics")
	{
		statisticsGroup.GET("/overview", statistics.Overview)
	}

	// 审计日志（需认证）
	auditGroup := api.Group("/audit")
	{
		auditGroup.GET("", audit.Index)
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
		file, err := staticFS.Open("index.html")
		if err != nil {
			logger.Errorf("读取首页文件失败: %s", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		c.Header("Content-Type", "text/html")
		_, _ = io.Copy(c.Writer, file)
	})

	// 静态文件路由 - 必须放在最后
	r.NoRoute(func(c *gin.Context) {
		filepath := c.Request.URL.Path

		// 移除 /public 前缀（如果存在）
		filepath = strings.TrimPrefix(filepath, "/public")
		filepath = strings.TrimPrefix(filepath, "/")

		// 尝试从 staticFS 读取文件
		file, err := staticFS.Open(filepath)
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
			_, _ = io.Copy(c.Writer, file)
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
	r.Use(securityHeaders)
	r.Use(checkAppInstall)
	r.Use(ipAuth)
	r.Use(userAuth)
	r.Use(urlAuth)
	r.Use(auditLog)
}

// securityHeaders 设置通用的安全响应头，防御点击劫持 / MIME sniff / referrer 泄漏。
// 不设置 CSP（需要针对前端资源单独调校）也不设置 HSTS（由反向代理决定）。
func securityHeaders(c *gin.Context) {
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Referrer-Policy", "no-referrer")
	c.Next()
}

// region Custom middleware

// isStaticFileRequest checks if the request is for a static file (non-API path).
// Static files are served via NoRoute handler and never match registered API routes.
func isStaticFileRequest(path string) bool {
	return !strings.HasPrefix(path, urlPrefix+"/") && !strings.HasPrefix(path, "/v1/")
}

// checkAppInstall verifies the application has been installed.
func checkAppInstall(c *gin.Context) {
	if app.Installed {
		c.Next()
		return
	}
	path := c.Request.URL.Path
	// Allow install API, root page, and static files before installation
	if strings.HasPrefix(path, "/api/install") || isStaticFileRequest(path) {
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
	clientIp := utils.ClientIP(c)
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

// userAuth authenticates the user for API requests.
func userAuth(c *gin.Context) {
	if !app.Installed {
		c.Next()
		return
	}

	path := c.Request.URL.Path
	// Static files (non-API paths) don't require authentication
	if isStaticFileRequest(path) {
		c.Next()
		return
	}

	uri := strings.TrimRight(path, "/")
	// 登录接口和安装状态接口不需要认证
	excludePaths := []string{"", "/api/user/login", "/api/install/status", "/api/agent/install.sh", "/api/agent/register", "/api/agent/download"}
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
	newToken, err := user.RestoreToken(c)
	if err != nil {
		logger.Warnf("token解析失败: %v, path: %s", err, path)
		jsonResp := utils.JsonResponse{}
		data := jsonResp.Failure(utils.AuthError, i18n.T(c, "auth_failed"))
		c.String(http.StatusOK, data)
		c.Abort()
		return
	}
	// 如果token被刷新，返回新token给前端
	if newToken != "" {
		c.Header("New-Auth-Token", newToken)
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

// urlAuth checks URL-level permissions (admin vs normal user).
func urlAuth(c *gin.Context) {
	if !app.Installed {
		c.Next()
		return
	}

	path := c.Request.URL.Path
	// Static files (non-API paths) don't require permission checks
	if isStaticFileRequest(path) {
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
		"/api/task/tags",
		"/api/task/log",
		"/api/host",
		"/api/host/all",
		"/api/user/login",
		"/api/user/editMyPassword",
		"/api/user/2fa/status",
		"/api/user/2fa/setup",
		"/api/user/2fa/enable",
		"/api/user/2fa/disable",
		"/api/template",
		"/api/template/categories",
		"/api/statistics/overview",
		"/api/agent/install.sh",
		"/api/agent/register",
		"/api/agent/download",
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

// auditLog middleware records audit log entries for write operations.
// It runs after the handler (post-processing) and only records successful POST requests.
func auditLog(c *gin.Context) {
	c.Next()

	// Only record POST requests
	if c.Request.Method != http.MethodPost {
		return
	}

	// Only record successful operations (status < 400)
	if c.Writer.Status() >= 400 {
		return
	}

	path := c.FullPath()
	username := user.Username(c)
	ip := utils.ClientIP(c)

	module, action := resolveModuleAction(path, c)
	if module == "" || action == "" {
		return
	}

	// 获取 targetId：优先从 URL 参数，其次从 POST body
	targetId := 0
	if idStr := c.Param("id"); idStr != "" {
		targetId, _ = strconv.Atoi(idStr)
	} else if idStr := c.PostForm("id"); idStr != "" && idStr != "0" {
		targetId, _ = strconv.Atoi(idStr)
	}

	// 读取 handler 设置的审计详情
	detail, _ := c.Get("audit_detail")
	detailStr, _ := detail.(string)

	log := &models.AuditLog{
		Username: username,
		Ip:       ip,
		Module:   module,
		Action:   action,
		TargetId: targetId,
		Detail:   detailStr,
	}

	// 异步查询对象名称并写入；使用独立 context 避免请求已结束后 goroutine 无界堆积
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.TargetName = resolveTargetName(ctx, module, targetId)
		if err := models.Db.WithContext(ctx).Create(log).Error; err != nil {
			logger.Warnf("写入审计日志失败: %v", err)
		}
	}()
}

// resolveModuleAction maps a Gin full path pattern to (module, action).
func resolveModuleAction(path string, c *gin.Context) (module, action string) {
	switch path {
	// Task routes
	case "/api/task/store":
		idStr := c.PostForm("id")
		if idStr == "" || idStr == "0" {
			return "task", "create"
		}
		return "task", "update"
	case "/api/task/remove/:id":
		return "task", "delete"
	case "/api/task/enable/:id":
		return "task", "enable"
	case "/api/task/disable/:id":
		return "task", "disable"
	case "/api/task/batch-enable":
		return "task", "batch-enable"
	case "/api/task/batch-disable":
		return "task", "batch-disable"
	case "/api/task/batch-remove":
		return "task", "batch-remove"

	// Host routes
	case "/api/host/store":
		idStr := c.PostForm("id")
		if idStr == "" || idStr == "0" {
			return "host", "create"
		}
		return "host", "update"
	case "/api/host/remove/:id":
		return "host", "delete"

	// User routes
	case "/api/user/store":
		idStr := c.PostForm("id")
		if idStr == "" || idStr == "0" {
			return "user", "create"
		}
		return "user", "update"
	case "/api/user/remove/:id":
		return "user", "delete"
	case "/api/user/enable/:id":
		return "user", "enable"
	case "/api/user/disable/:id":
		return "user", "disable"
	case "/api/user/editMyPassword":
		return "user", "change-password"
	case "/api/user/editPassword/:id":
		return "user", "reset-password"

	// Template routes
	case "/api/template/store":
		idStr := c.PostForm("id")
		if idStr == "" || idStr == "0" {
			return "template", "create"
		}
		return "template", "update"
	case "/api/template/remove/:id":
		return "template", "delete"
	case "/api/template/apply/:id":
		return "template", "update"
	case "/api/template/save-from-task":
		return "template", "create"

	// System routes — any POST under /api/system
	default:
		if strings.HasPrefix(path, "/api/system/") {
			return "system", "update"
		}
	}

	return "", ""
}

// resolveTargetName 根据 module 和 targetId 查询对象名称
func resolveTargetName(ctx context.Context, module string, targetId int) string {
	if targetId == 0 {
		return ""
	}
	db := models.Db.WithContext(ctx)
	switch module {
	case "task":
		task := &models.Task{}
		if err := db.Select("name").First(task, targetId).Error; err == nil {
			return task.Name
		}
	case "host":
		host := &models.Host{}
		if err := db.Select("name", "alias").First(host, targetId).Error; err == nil {
			if host.Alias != "" {
				return host.Alias
			}
			return host.Name
		}
	case "user":
		u := &models.User{}
		if err := db.Select("name").First(u, targetId).Error; err == nil {
			return u.Name
		}
	case "template":
		tmpl := &models.TaskTemplate{}
		if err := models.Db.Select("name").First(tmpl, targetId).Error; err == nil {
			return tmpl.Name
		}
	}
	return ""
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
	realSign := utils.Sha256(raw)
	if subtle.ConstantTimeCompare([]byte(sign), []byte(realSign)) != 1 {
		msg := json.CommonFailure(i18n.T(c, "sign_verify_failed"))
		c.String(http.StatusOK, msg)
		c.Abort()
		return
	}
	c.Next()
}

// endregion
