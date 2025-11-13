package models

import (
	"errors"

	"github.com/gocronx-team/gocron/internal/modules/logger"
	"gorm.io/gorm"
)

type Migration struct{}

// 首次安装, 创建数据库表
func (migration *Migration) Install(dbName string) error {
	setting := new(Setting)
	tables := []interface{}{
		&User{}, &Task{}, &TaskLog{}, &Host{}, setting, &LoginLog{}, &TaskHost{}, &AgentToken{},
	}

	for _, table := range tables {
		if Db.Migrator().HasTable(table) {
			return errors.New("数据表已存在")
		}
		err := Db.AutoMigrate(table)
		if err != nil {
			return err
		}
	}

	// SQLite特殊处理：修复task_log表的自增主键
	if Db.Dialector.Name() == "sqlite" {
		migration.fixSQLiteAutoIncrement()
	}

	setting.InitBasicField()

	return nil
}

// 迭代升级数据库, 新建表、新增字段等
func (migration *Migration) Upgrade(oldVersionId int) {
	// v1.2版本不支持升级
	if oldVersionId == 120 {
		return
	}

	versionIds := []int{110, 122, 130, 140, 150, 151, 152, 153, 154}
	upgradeFuncs := []func(*gorm.DB) error{
		migration.upgradeFor110,
		migration.upgradeFor122,
		migration.upgradeFor130,
		migration.upgradeFor140,
		migration.upgradeFor150,
		migration.upgradeFor151,
		migration.upgradeFor152,
		migration.upgradeFor153,
		migration.upgradeFor154,
	}

	startIndex := -1
	// 从当前版本的下一版本开始升级
	for i, value := range versionIds {
		if value > oldVersionId {
			startIndex = i
			break
		}
	}

	if startIndex == -1 {
		return
	}

	length := len(versionIds)
	if startIndex >= length {
		return
	}

	err := Db.Transaction(func(tx *gorm.DB) error {
		for startIndex < length {
			err := upgradeFuncs[startIndex](tx)
			if err != nil {
				return err
			}
			startIndex++
		}
		return nil
	})

	if err != nil {
		logger.Fatal("数据库升级失败", err)
	}
}

// 升级到v1.1版本
func (migration *Migration) upgradeFor110(tx *gorm.DB) error {
	logger.Info("开始升级到v1.1")

	// 创建表task_host
	err := tx.AutoMigrate(&TaskHost{})
	if err != nil {
		return err
	}

	// 把task对应的host_id写入task_host表
	type OldTask struct {
		Id     int
		HostId int16
	}
	var results []OldTask
	err = tx.Table(TablePrefix+"task").Select("id", "host_id").Where("host_id > ?", 0).Find(&results).Error
	if err != nil {
		return err
	}

	for _, value := range results {
		taskHostModel := &TaskHost{
			TaskId: value.Id,
			HostId: value.HostId,
		}
		err = tx.Create(taskHostModel).Error
		if err != nil {
			return err
		}
	}

	// 删除task表host_id字段
	err = tx.Migrator().DropColumn(&Task{}, "host_id")

	logger.Info("已升级到v1.1\n")

	return err
}

// 升级到1.2.2版本
func (migration *Migration) upgradeFor122(tx *gorm.DB) error {
	logger.Info("开始升级到v1.2.2")

	// task表增加tag字段
	if !tx.Migrator().HasColumn(&Task{}, "tag") {
		err := tx.Migrator().AddColumn(&Task{}, "tag")
		if err != nil {
			return err
		}
	}

	logger.Info("已升级到v1.2.2\n")

	return nil
}

// 升级到v1.3版本
func (migration *Migration) upgradeFor130(tx *gorm.DB) error {
	logger.Info("开始升级到v1.3")

	// 删除user表deleted字段（如果存在）
	if tx.Migrator().HasColumn(&User{}, "deleted") {
		err := tx.Migrator().DropColumn(&User{}, "deleted")
		if err != nil {
			return err
		}
	}

	logger.Info("已升级到v1.3\n")

	return nil
}

// 升级到v1.4版本
func (migration *Migration) upgradeFor140(tx *gorm.DB) error {
	logger.Info("开始升级到v1.4")

	// task表增加字段
	// retry_interval 重试间隔时间(秒)
	// http_method    http请求方法
	if !tx.Migrator().HasColumn(&Task{}, "retry_interval") {
		err := tx.Migrator().AddColumn(&Task{}, "retry_interval")
		if err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Task{}, "http_method") {
		err := tx.Migrator().AddColumn(&Task{}, "http_method")
		if err != nil {
			return err
		}
	}

	logger.Info("已升级到v1.4\n")

	return nil
}

func (m *Migration) upgradeFor150(tx *gorm.DB) error {
	logger.Info("开始升级到v1.5")

	// task表增加字段 notify_keyword
	if !tx.Migrator().HasColumn(&Task{}, "notify_keyword") {
		err := tx.Migrator().AddColumn(&Task{}, "notify_keyword")
		if err != nil {
			return err
		}
	}

	// 检查并创建邮件模板配置
	var count int64
	tx.Model(&Setting{}).Where("code = ? AND key = ?", MailCode, MailTemplateKey).Count(&count)
	if count == 0 {
		settingModel := &Setting{
			Code:  MailCode,
			Key:   MailTemplateKey,
			Value: emailTemplate,
		}
		if err := tx.Create(settingModel).Error; err != nil {
			return err
		}
	}

	// 检查并创建Slack模板配置
	tx.Model(&Setting{}).Where("code = ? AND key = ?", SlackCode, SlackTemplateKey).Count(&count)
	if count == 0 {
		settingModel := &Setting{
			Code:  SlackCode,
			Key:   SlackTemplateKey,
			Value: slackTemplate,
		}
		if err := tx.Create(settingModel).Error; err != nil {
			return err
		}
	}

	// 检查并创建Webhook URL配置
	tx.Model(&Setting{}).Where("code = ? AND key = ?", WebhookCode, WebhookUrlKey).Count(&count)
	if count == 0 {
		settingModel := &Setting{
			Code:  WebhookCode,
			Key:   WebhookUrlKey,
			Value: "",
		}
		if err := tx.Create(settingModel).Error; err != nil {
			return err
		}
	}

	// 检查并创建Webhook模板配置
	tx.Model(&Setting{}).Where("code = ? AND key = ?", WebhookCode, WebhookTemplateKey).Count(&count)
	if count == 0 {
		settingModel := &Setting{
			Code:  WebhookCode,
			Key:   WebhookTemplateKey,
			Value: webhookTemplate,
		}
		if err := tx.Create(settingModel).Error; err != nil {
			return err
		}
	}

	logger.Info("已升级到v1.5\n")

	return nil
}

// 升级到v1.5.1版本 - 添加2FA字段
func (m *Migration) upgradeFor151(tx *gorm.DB) error {
	logger.Info("开始升级到v1.5.1 - 添加2FA支持")

	// user表增加two_factor_key字段
	if !tx.Migrator().HasColumn(&User{}, "two_factor_key") {
		err := tx.Migrator().AddColumn(&User{}, "two_factor_key")
		if err != nil {
			return err
		}
	}

	// user表增加two_factor_on字段
	if !tx.Migrator().HasColumn(&User{}, "two_factor_on") {
		err := tx.Migrator().AddColumn(&User{}, "two_factor_on")
		if err != nil {
			return err
		}
	}

	logger.Info("已升级到v1.5.1\n")

	return nil
}

// 升级到v1.5.2版本 - 修复 SQLite host 表 AUTOINCREMENT
func (m *Migration) upgradeFor152(tx *gorm.DB) error {
	logger.Info("开始升级到v1.5.2 - 修复 host 表自增主键")

	// 只对 SQLite 数据库执行修复
	if tx.Dialector.Name() == "sqlite" {
		var tableSQL string
		err := tx.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='host'").Scan(&tableSQL).Error
		if err != nil {
			return err
		}

		if len(tableSQL) > 0 && !contains(tableSQL, "AUTOINCREMENT") {
			logger.Info("检测到 host 表需要修复")

			// 检查是否有数据
			var hasData int64
			tx.Raw("SELECT COUNT(*) FROM host").Scan(&hasData)

			// 重建表以支持 AUTOINCREMENT
			err = tx.Exec(`
				CREATE TABLE IF NOT EXISTS host_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name varchar(64) NOT NULL,
					alias varchar(32) NOT NULL DEFAULT '',
					port integer NOT NULL DEFAULT 5921,
					remark varchar(100) NOT NULL DEFAULT ''
				);
			`).Error
			if err != nil {
				return err
			}

			// 如果有数据，迁移数据
			if hasData > 0 {
				err = tx.Exec(`
					INSERT INTO host_new (name, alias, port, remark)
					SELECT name, alias, port, remark FROM host WHERE name IS NOT NULL;
				`).Error
				if err != nil {
					return err
				}
			}

			// 删除旧表
			err = tx.Exec(`DROP TABLE host;`).Error
			if err != nil {
				return err
			}

			// 重命名新表
			err = tx.Exec(`ALTER TABLE host_new RENAME TO host;`).Error
			if err != nil {
				return err
			}

			logger.Info("host 表已重建，支持自增主键")
		} else {
			logger.Info("host 表结构正确，无需修复")
		}
	}

	logger.Info("已升级到v1.5.2\n")

	return nil
}

// 升级到v1.5.3版本 - 修复 SQLite task_log 表 AUTOINCREMENT
func (m *Migration) upgradeFor153(tx *gorm.DB) error {
	logger.Info("开始升级到v1.5.3 - 修复 task_log 表自增主键")

	// 只对 SQLite 数据库执行修复
	if tx.Dialector.Name() == "sqlite" {
		var tableSQL string
		err := tx.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='task_log'").Scan(&tableSQL).Error
		if err != nil {
			return err
		}

		if len(tableSQL) > 0 && !contains(tableSQL, "AUTOINCREMENT") {
			logger.Info("检测到 task_log 表需要修复")

			err = tx.Exec(`
				CREATE TABLE IF NOT EXISTS task_log_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					task_id integer NOT NULL DEFAULT 0,
					name varchar(32) NOT NULL,
					spec varchar(64) NOT NULL,
					protocol tinyint NOT NULL,
					command varchar(256) NOT NULL,
					timeout mediumint NOT NULL DEFAULT 0,
					retry_times tinyint NOT NULL DEFAULT 0,
					hostname varchar(128) NOT NULL DEFAULT '',
					start_time datetime,
					end_time datetime,
					status tinyint NOT NULL DEFAULT 1,
					result mediumtext NOT NULL
				);
			`).Error
			if err != nil {
				return err
			}

			// 迁移最近的数据（最多10000条）
			var hasData int64
			tx.Raw("SELECT COUNT(*) FROM task_log").Scan(&hasData)
			if hasData > 0 {
				err = tx.Exec(`
					INSERT INTO task_log_new (task_id, name, spec, protocol, command, timeout, retry_times, hostname, start_time, end_time, status, result)
					SELECT task_id, name, spec, protocol, command, timeout, retry_times, hostname, start_time, end_time, status, result 
					FROM task_log 
					WHERE task_id IS NOT NULL
					ORDER BY start_time DESC 
					LIMIT 10000;
				`).Error
				if err != nil {
					return err
				}
			}

			err = tx.Exec(`DROP TABLE task_log;`).Error
			if err != nil {
				return err
			}

			err = tx.Exec(`ALTER TABLE task_log_new RENAME TO task_log;`).Error
			if err != nil {
				return err
			}

			logger.Info("task_log 表已重建，支持自增主键")
		} else {
			logger.Info("task_log 表结构正确，无需修复")
		}

		// 清理状态异常的历史任务日志（status=1 且 result 为空）
		err = tx.Exec(`
			UPDATE task_log 
			SET status = 0, 
			    result = '任务异常终止（未正常完成）',
			    end_time = datetime(start_time, '+1 second')
			WHERE status = 1 
			AND (result IS NULL OR result = '');
		`).Error
		if err != nil {
			logger.Error("清理异常任务日志失败", err)
		} else {
			logger.Info("已清理状态异常的历史任务日志")
		}
	}

	logger.Info("已升级到v1.5.3\n")

	return nil
}

// 升级到v1.5.4版本 - 添加agent_token表
func (m *Migration) upgradeFor154(tx *gorm.DB) error {
	logger.Info("开始升级到v1.5.4 - 添加agent自动注册支持")

	if err := tx.AutoMigrate(&AgentToken{}); err != nil {
		return err
	}

	if err := tx.Migrator().AlterColumn(&AgentToken{}, "UsedAt"); err != nil {
		logger.Warn("调整 agent_token.used_at 可空属性失败", err)
	}

	logger.Info("已升级到v1.5.4\n")

	return nil
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 修复SQLite表的自增主键问题
func (m *Migration) fixSQLiteAutoIncrement() {
	logger.Info("检查SQLite表结构...")

	// 修复task_log表
	var taskLogSQL string
	Db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='task_log'").Scan(&taskLogSQL)
	if len(taskLogSQL) > 0 && !contains(taskLogSQL, "AUTOINCREMENT") {
		logger.Info("修复task_log表自增主键...")
		Db.Exec(`
			CREATE TABLE task_log_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				task_id integer NOT NULL DEFAULT 0,
				name varchar(32) NOT NULL,
				spec varchar(64) NOT NULL,
				protocol tinyint NOT NULL,
				command varchar(256) NOT NULL,
				timeout mediumint NOT NULL DEFAULT 0,
				retry_times tinyint NOT NULL DEFAULT 0,
				hostname varchar(128) NOT NULL DEFAULT '',
				start_time datetime,
				end_time datetime,
				status tinyint NOT NULL DEFAULT 1,
				result mediumtext NOT NULL
			);
		`)
		Db.Exec(`DROP TABLE task_log;`)
		Db.Exec(`ALTER TABLE task_log_new RENAME TO task_log;`)
		logger.Info("修复task_log表完成")
	}

	// 修复host表
	var hostSQL string
	Db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='host'").Scan(&hostSQL)
	if len(hostSQL) > 0 && !contains(hostSQL, "AUTOINCREMENT") {
		logger.Info("修复host表自增主键...")
		Db.Exec(`
			CREATE TABLE host_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name varchar(64) NOT NULL,
				alias varchar(32) NOT NULL DEFAULT '',
				port integer NOT NULL DEFAULT 5921,
				remark varchar(100) NOT NULL DEFAULT ''
			);
		`)
		Db.Exec(`DROP TABLE host;`)
		Db.Exec(`ALTER TABLE host_new RENAME TO host;`)
		logger.Info("修复host表完成")
	}
}
