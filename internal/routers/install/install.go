package install

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/service"
	"github.com/lib/pq"
)

// 系统安装

type InstallForm struct {
	DbType               string `form:"db_type" binding:"required,oneof=mysql postgres sqlite"`
	DbHost               string `form:"db_host" binding:"max=50"`
	DbPort               int    `form:"db_port" binding:"min=0,max=65535"`
	DbUsername           string `form:"db_username" binding:"max=50"`
	DbPassword           string `form:"db_password" binding:"max=30"`
	DbName               string `form:"db_name" binding:"required,max=200"`
	DbTablePrefix        string `form:"db_table_prefix" binding:"max=20"`
	AdminUsername        string `form:"admin_username" binding:"required,min=3"`
	AdminPassword        string `form:"admin_password" binding:"required,min=6"`
	ConfirmAdminPassword string `form:"confirm_admin_password" binding:"required,min=6"`
	AdminEmail           string `form:"admin_email" binding:"required,email,max=50"`
}

// 安装
func Store(c *gin.Context) {
	var form InstallForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondError(c, "表单验证失败, 请检测输入")
		return
	}

	if app.Installed {
		base.RespondError(c, "系统已安装!")
		return
	}
	if form.AdminPassword != form.ConfirmAdminPassword {
		base.RespondError(c, "两次输入密码不匹配")
		return
	}
	err := testDbConnection(form)
	if err != nil {
		base.RespondError(c, err.Error())
		return
	}
	// 写入数据库配置
	err = writeConfig(form)
	if err != nil {
		base.RespondError(c, "数据库配置写入文件失败", err)
		return
	}

	appConfig, err := setting.Read(app.AppConfig)
	if err != nil {
		base.RespondError(c, "读取应用配置失败", err)
		return
	}
	app.Setting = appConfig

	models.Db = models.CreateDb()
	// 创建数据库表
	migration := new(models.Migration)
	err = migration.Install(form.DbName)
	if err != nil {
		base.RespondError(c, fmt.Sprintf("创建数据库表失败-%s", err.Error()), err)
		return
	}

	// 创建管理员账号
	err = createAdminUser(form)
	if err != nil {
		base.RespondError(c, "创建管理员账号失败", err)
		return
	}

	// 创建安装锁
	err = app.CreateInstallLock()
	if err != nil {
		base.RespondError(c, "创建文件安装锁失败", err)
		return
	}

	// 更新版本号文件
	app.UpdateVersionFile()

	// 标记为已安装
	app.Installed = true
	// 初始化并启动定时任务调度器
	service.ServiceTask.Initialize()
	service.ServiceTask.StartScheduler()

	base.RespondSuccess(c, "安装成功", nil)
}

// 配置写入文件
func writeConfig(form InstallForm) error {
	dbHost := form.DbHost
	dbPort := strconv.Itoa(form.DbPort)
	if form.DbType == "sqlite" {
		dbHost = ""
		dbPort = "0"
	}
	dbConfig := []string{
		"db.engine", form.DbType,
		"db.host", dbHost,
		"db.port", dbPort,
		"db.user", form.DbUsername,
		"db.password", form.DbPassword,
		"db.database", form.DbName,
		"db.prefix", form.DbTablePrefix,
		"db.charset", "utf8",
		"db.max.idle.conns", "5",
		"db.max.open.conns", "100",
		"allow_ips", "",
		"app.name", "定时任务管理系统", // 应用名称
		"api.key", "",
		"api.secret", "",
		"enable_tls", "false",
		"concurrency.queue", "500",
		"auth_secret", utils.RandAuthToken(),
		"ca_file", "",
		"cert_file", "",
		"key_file", "",
	}

	return setting.Write(dbConfig, app.AppConfig)
}

// 创建管理员账号
func createAdminUser(form InstallForm) error {
	user := new(models.User)
	user.Name = form.AdminUsername
	user.Password = form.AdminPassword
	user.Email = form.AdminEmail
	user.IsAdmin = 1
	_, err := user.Create()

	return err
}

// 测试数据库连接
func testDbConnection(form InstallForm) error {
	var s setting.Setting
	s.Db.Engine = form.DbType
	s.Db.Host = form.DbHost
	s.Db.Port = form.DbPort
	s.Db.User = form.DbUsername
	s.Db.Password = form.DbPassword
	s.Db.Database = form.DbName
	s.Db.Charset = "utf8"

	// SQLite 不需要测试连接，会自动创建文件
	if s.Db.Engine == "sqlite" {
		return nil
	}

	db, err := models.CreateTmpDb(&s)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	err = sqlDB.Ping()
	if s.Db.Engine == "postgres" && err != nil {
		pgError, ok := err.(*pq.Error)
		if ok && pgError.Code == "3D000" {
			err = errors.New("数据库不存在")
		}
		return err
	}

	if s.Db.Engine == "mysql" && err != nil {
		mysqlError, ok := err.(*mysql.MySQLError)
		if ok && mysqlError.Number == 1049 {
			err = errors.New("数据库不存在")
		}
		return err
	}

	return err

}
