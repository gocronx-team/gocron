package app

import (
	"os"
	"path/filepath"

	"fmt"
	"strconv"
	"strings"

	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"github.com/gocronx-team/gocron/internal/modules/utils"
)

var (
	// AppDir 应用根目录
	AppDir string // 应用根目录
	// ConfDir 配置文件目录
	ConfDir string // 配置目录
	// LogDir 日志目录
	LogDir string // 日志目录
	// AppConfig 配置文件
	AppConfig string // 应用配置文件
	// Installed 应用是否已安装
	Installed bool // 应用是否安装过
	// Setting 应用配置
	Setting *setting.Setting // 应用配置
	// VersionId 版本号
	VersionId int // 版本号
	// VersionFile 版本文件
	VersionFile string // 版本号文件
)

// InitEnv 初始化
func InitEnv(versionString string) {
	logger.InitLogger()
	var err error
	// 开发环境使用当前目录，生产环境使用可执行文件目录
	execPath, err := os.Executable()
	if err != nil {
		logger.Fatal(err)
	}
	execDir := filepath.Dir(execPath)

	// 开发环境检测：Air 热重载（tmp 目录）或 go run（go-build cache 目录）
	execName := filepath.Base(execPath)
	if filepath.Base(execDir) == "tmp" {
		AppDir = filepath.Join(filepath.Dir(execDir), ".gocron")
	} else if strings.Contains(execDir, "go-build") && !strings.HasSuffix(execName, ".test") {
		// go run 会将二进制编译到 go-build cache 中，使用当前工作目录
		// 排除 go test（测试二进制以 .test 结尾）
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			logger.Fatal(wdErr)
		}
		AppDir = filepath.Join(wd, ".gocron")
	} else {
		AppDir = filepath.Join(execDir, ".gocron")
	}
	fmt.Printf("AppDir: %s\n", AppDir)
	ConfDir = filepath.Join(AppDir, "conf")
	LogDir = filepath.Join(AppDir, "log")
	AppConfig = filepath.Join(ConfDir, "app.ini")
	VersionFile = filepath.Join(ConfDir, ".version")
	fmt.Printf("ConfDir: %s, LogDir: %s\n", ConfDir, LogDir)
	createDirIfNotExists(AppDir, ConfDir, LogDir)
	Installed = IsInstalled()
	VersionId = ToNumberVersion(versionString)
}

// IsInstalled 判断应用是否已安装
func IsInstalled() bool {
	_, err := os.Stat(filepath.Join(ConfDir, "install.lock"))
	return !os.IsNotExist(err)
}

// CreateInstallLock 创建安装锁文件
func CreateInstallLock() error {
	lockFile := filepath.Join(ConfDir, "install.lock")
	err := os.WriteFile(lockFile, []byte(""), 0600)
	if err != nil {
		logger.Error("创建安装锁文件conf/install.lock失败", err)
		fmt.Printf("Error creating install.lock: %v\n", err)
	} else {
		fmt.Printf("Successfully created install.lock at %s\n", lockFile)
	}

	return err
}

// UpdateVersionFile 更新应用版本号文件
func UpdateVersionFile() {
	err := os.WriteFile(VersionFile,
		[]byte(strconv.Itoa(VersionId)),
		0600,
	)

	if err != nil {
		logger.Fatal(err)
	}
}

// GetCurrentVersionId 获取应用当前版本号, 从版本号文件中读取
func GetCurrentVersionId() int {
	if !utils.FileExist(VersionFile) {
		return 0
	}

	bytes, err := os.ReadFile(VersionFile)
	if err != nil {
		logger.Fatal(err)
	}

	versionId, err := strconv.Atoi(strings.TrimSpace(string(bytes)))
	if err != nil {
		logger.Fatal(err)
	}

	return versionId
}

// ToNumberVersion 把字符串版本号a.b.c转换为整数版本号abc
func ToNumberVersion(versionString string) int {
	versionString = strings.TrimPrefix(versionString, "v")
	v := strings.Replace(versionString, ".", "", -1)
	if len(v) < 3 {
		v += "0"
	}

	versionId, err := strconv.Atoi(v)
	if err != nil {
		logger.Fatal(err)
	}

	return versionId
}

// 检测目录是否存在
func createDirIfNotExists(path ...string) {
	for _, value := range path {
		if utils.FileExist(value) {
			continue
		}
		err := os.MkdirAll(value, 0755)
		if err != nil {
			logger.Fatal(fmt.Sprintf("创建目录失败:%s", err.Error()))
		}
	}
}
