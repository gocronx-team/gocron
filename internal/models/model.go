package models

import (
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/gocronx-team/gocron/internal/modules/app"
	glogger "github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/setting"
)

type Status int8
type CommonMap map[string]interface{}

var TablePrefix = ""
var Db *gorm.DB

// dbKeepAliveStop is closed to signal the keepDbAlived goroutine to exit.
var dbKeepAliveStop chan struct{}

const (
	Disabled Status = 0 // 禁用
	Failure  Status = 0 // 失败
	Enabled  Status = 1 // 启用
	Running  Status = 1 // 运行中
	Finish   Status = 2 // 完成
	Cancel   Status = 3 // 取消
)

const (
	Page        = 1    // 当前页数
	PageSize    = 20   // 每页多少条数据
	MaxPageSize = 1000 // 每次最多取多少条
)

const DefaultTimeFormat = "2006-01-02 15:04:05"

const (
	dbPingInterval = 90 * time.Second
	dbMaxLiftTime  = 2 * time.Hour
)

type BaseModel struct {
	Page     int `gorm:"-"`
	PageSize int `gorm:"-"`
}

func (model *BaseModel) parsePageAndPageSize(params CommonMap) {
	page, ok := params["Page"]
	if ok {
		model.Page = page.(int)
	}
	pageSize, ok := params["PageSize"]
	if ok {
		model.PageSize = pageSize.(int)
	}
	if model.Page <= 0 {
		model.Page = Page
	}
	if model.PageSize <= 0 {
		model.PageSize = MaxPageSize
	}
}

func (model *BaseModel) pageLimitOffset() int {
	return (model.Page - 1) * model.PageSize
}

// 创建Db
func CreateDb() *gorm.DB {
	dsn := getDbEngineDSN(app.Setting)
	var dialector gorm.Dialector

	engine := strings.ToLower(app.Setting.Db.Engine)
	switch engine {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		ensureSqliteDir(dsn)
		dialector = gormlite.Open(dsn)
	default:
		glogger.Fatal("不支持的数据库类型", nil)
	}

	// 配置 gorm
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   app.Setting.Db.Prefix,
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// 开发模式下开启日志
	if gin.Mode() == gin.DebugMode {
		config.Logger = logger.New(
			stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags),
			logger.Config{
				SlowThreshold:        200 * time.Millisecond,
				LogLevel:             logger.Info,
				Colorful:             true,
				ParameterizedQueries: true,
			},
		)
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		glogger.Fatal("创建gorm引擎失败", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		glogger.Fatal("获取数据库连接失败", err)
	}

	// SQLite 需要特殊的连接池配置
	if engine == "sqlite" {
		sqlDB.SetMaxOpenConns(1) // SQLite 只允许一个写连接
	} else {
		sqlDB.SetMaxIdleConns(app.Setting.Db.MaxIdleConns)
		sqlDB.SetMaxOpenConns(app.Setting.Db.MaxOpenConns)
	}
	sqlDB.SetConnMaxLifetime(dbMaxLiftTime)

	if app.Setting.Db.Prefix != "" {
		TablePrefix = app.Setting.Db.Prefix
	}

	dbKeepAliveStop = make(chan struct{})
	go keepDbAlived(db, dbKeepAliveStop)

	return db
}

// StopKeepAlive signals the keepDbAlived goroutine to exit. Safe to call multiple times.
func StopKeepAlive() {
	if dbKeepAliveStop == nil {
		return
	}
	select {
	case <-dbKeepAliveStop:
		// already closed
	default:
		close(dbKeepAliveStop)
	}
}

// 创建临时数据库连接
func CreateTmpDb(setting *setting.Setting) (*gorm.DB, error) {
	dsn := getDbEngineDSN(setting)
	var dialector gorm.Dialector

	engine := strings.ToLower(setting.Db.Engine)
	switch engine {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		ensureSqliteDir(dsn)
		dialector = gormlite.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", engine)
	}

	return gorm.Open(dialector, &gorm.Config{})
}

// 获取数据库引擎DSN  mysql,postgres
func getDbEngineDSN(setting *setting.Setting) string {
	engine := strings.ToLower(setting.Db.Engine)
	dsn := ""
	switch engine {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			setting.Db.User,
			setting.Db.Password,
			setting.Db.Host,
			setting.Db.Port,
			setting.Db.Database,
			setting.Db.Charset)
	case "postgres":
		dsn = fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
			setting.Db.User,
			setting.Db.Password,
			setting.Db.Host,
			setting.Db.Port,
			setting.Db.Database)
	case "sqlite":
		dsn = setting.Db.Database
	}

	return dsn
}

func keepDbAlived(db *gorm.DB, stop <-chan struct{}) {
	ticker := time.NewTicker(dbPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			sqlDB, err := db.DB()
			if err != nil {
				glogger.Infof("database get connection: %s", err)
				continue
			}
			if err := sqlDB.Ping(); err != nil {
				glogger.Infof("database ping failed: %s", err)
			} else {
				glogger.Infof("database ping: ok")
			}
		}
	}
}

// 确保 SQLite 数据库文件所在目录存在
func ensureSqliteDir(dbPath string) {
	// 清理并规范化路径
	dbPath = filepath.Clean(dbPath)
	dir := filepath.Dir(dbPath)

	if dir != "" && dir != "." {
		// 验证路径不是绝对路径时，确保不包含父目录引用
		if !filepath.IsAbs(dbPath) && strings.Contains(dbPath, "..") {
			glogger.Fatal("非法的数据库路径", nil)
		}
		if err := os.MkdirAll(dir, 0750); err != nil {
			glogger.Fatal("创建SQLite数据库目录失败", err)
		}
	}
}
