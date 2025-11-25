// Command gocron
//go:generate statik -src=../../web/public -dest=../../internal -f

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers"
	"github.com/gocronx-team/gocron/internal/service"
	"github.com/urfave/cli"
)

var (
	AppVersion           = "1.5.3"
	BuildDate, GitCommit string
)

// web服务器默认端口
const DefaultPort = 5920

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "gocron"
	cliApp.Usage = "gocron service"
	cliApp.Version, _ = utils.FormatAppVersion(AppVersion, GitCommit, BuildDate)
	cliApp.Commands = getCommands()
	cliApp.Flags = append(cliApp.Flags, []cli.Flag{}...)

	// Windows下双击运行时自动添加web命令
	if len(os.Args) == 1 && utils.IsWindows() {
		os.Args = append(os.Args, "web")
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}

// getCommands
func getCommands() []cli.Command {
	command := cli.Command{
		Name:   "web",
		Usage:  "run web server",
		Action: runWeb,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "host",
				Value: "0.0.0.0",
				Usage: "bind host",
			},
			cli.IntFlag{
				Name:  "port,p",
				Value: DefaultPort,
				Usage: "bind port",
			},
			cli.StringFlag{
				Name:  "env,e",
				Value: "prod",
				Usage: "runtime environment, dev|test|prod",
			},
		},
	}

	return []cli.Command{command}
}

func runWeb(ctx *cli.Context) {
	// 设置运行环境
	setEnvironment(ctx)
	fmt.Printf("Starting gocron web server...\n")
	// 初始化应用
	app.InitEnv(AppVersion)
	fmt.Printf("Application initialized\n")
	// 初始化模块 DB、定时任务等
	initModule()
	fmt.Printf("Modules initialized\n")
	// 捕捉信号,配置热更新等
	go catchSignal()
	r := gin.Default()
	// 注册中间件
	routers.RegisterMiddleware(r)
	// 注册路由
	routers.Register(r)
	host := parseHost(ctx)
	port := parsePort(ctx)
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting server on %s\n", addr)
	err := r.Run(addr)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		logger.Fatal("Failed to start server", err)
	}
}

func initModule() {
	if !app.Installed {
		return
	}

	config, err := setting.Read(app.AppConfig)
	if err != nil {
		logger.Fatal("读取应用配置失败", err)
	}
	app.Setting = config

	// 初始化DB
	models.Db = models.CreateDb()

	// 版本升级
	upgradeIfNeed()

	// 自动创建缺失的表
	ensureTables()

	// 修复缺失的配置记录
	if err := models.RepairSettings(); err != nil {
		logger.Error("修复配置记录失败", err)
	}

	// 初始化定时任务
	service.ServiceTask.Initialize()
}

// 解析端口
func parsePort(ctx *cli.Context) int {
	port := DefaultPort
	if ctx.IsSet("port") {
		port = ctx.Int("port")
	}
	if port <= 0 || port >= 65535 {
		port = DefaultPort
	}

	return port
}

func parseHost(ctx *cli.Context) string {
	if ctx.IsSet("host") {
		return ctx.String("host")
	}

	return "0.0.0.0"
}

func setEnvironment(ctx *cli.Context) {
	env := "prod"
	if ctx.IsSet("env") {
		env = ctx.String("env")
	}

	switch env {
	case "test":
		gin.SetMode(gin.TestMode)
	case "dev":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

// 捕捉信号
func catchSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		logger.Info("收到信号 -- ", s)
		switch s {
		case syscall.SIGHUP:
			logger.Info("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			shutdown()
		}
	}
}

// 应用退出
func shutdown() {
	defer func() {
		logger.Info("已退出")
		logger.Close() // 确保日志刷新
		os.Exit(0)
	}()

	if !app.Installed {
		return
	}
	logger.Info("应用准备退出")
	// 停止所有任务调度
	logger.Info("停止定时任务调度")
	service.ServiceTask.WaitAndExit()
}

// 判断应用是否需要升级, 当存在版本号文件且版本小于app.VersionId时升级
func upgradeIfNeed() {
	currentVersionId := app.GetCurrentVersionId()
	// 没有版本号文件
	if currentVersionId == 0 {
		return
	}
	if currentVersionId >= app.VersionId {
		return
	}

	migration := new(models.Migration)
	logger.Infof("版本升级开始, 当前版本号%d", currentVersionId)

	migration.Upgrade(currentVersionId)
	app.UpdateVersionFile()

	logger.Infof("已升级到最新版本%d", app.VersionId)
}

// 确保所有表都存在
func ensureTables() {
	if !models.Db.Migrator().HasTable(&models.AgentToken{}) {
		logger.Info("检测到agent_token表不存在，开始创建...")
		if err := models.Db.AutoMigrate(&models.AgentToken{}); err != nil {
			logger.Error("创建agent_token表失败", err)
		} else {
			logger.Info("agent_token表创建成功")
		}
	}
}
