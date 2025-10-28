package models

import (
	"time"

	"github.com/gocronx-team/gocron/internal/modules/utils"
)

const PasswordSaltLength = 6

// 用户model
type User struct {
	Id           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"name" gorm:"type:varchar(32);not null;uniqueIndex"`
	Password     string    `json:"-" gorm:"type:char(32);not null"`
	Salt         string    `json:"-" gorm:"type:char(6);not null"`
	Email        string    `json:"email" gorm:"type:varchar(50);not null;uniqueIndex;default:''"`
	TwoFactorKey string    `json:"-" gorm:"column:two_factor_key;type:varchar(100);default:''"`
	TwoFactorOn  int8      `json:"two_factor_on" gorm:"column:two_factor_on;type:tinyint;not null;default:0"`
	CreatedAt    time.Time `json:"created" gorm:"column:created;autoCreateTime"`
	UpdatedAt    time.Time `json:"updated" gorm:"column:updated;autoUpdateTime"`
	IsAdmin      int8      `json:"is_admin" gorm:"type:tinyint;not null;default:0"`
	Status       Status    `json:"status" gorm:"type:tinyint;not null;default:1"`
	BaseModel    `json:"-" gorm:"-"`
}

// 新增
func (user *User) Create() (insertId int, err error) {
	user.Status = Enabled
	user.Salt = user.generateSalt()
	user.Password = user.encryptPassword(user.Password, user.Salt)

	result := Db.Create(user)
	if result.Error == nil {
		insertId = user.Id
	}

	return insertId, result.Error
}

// 更新
func (user *User) Update(id int, data CommonMap) (int64, error) {
	updateData := make(map[string]interface{})
	for k, v := range data {
		updateData[k] = v
	}
	result := Db.Model(&User{}).Where("id = ?", id).UpdateColumns(updateData)
	return result.RowsAffected, result.Error
}

func (user *User) UpdatePassword(id int, password string) (int64, error) {
	salt := user.generateSalt()
	safePassword := user.encryptPassword(password, salt)

	return user.Update(id, CommonMap{"password": safePassword, "salt": salt})
}

// 删除
func (user *User) Delete(id int) (int64, error) {
	result := Db.Delete(&User{}, id)
	return result.RowsAffected, result.Error
}

// 禁用
func (user *User) Disable(id int) (int64, error) {
	return user.Update(id, CommonMap{"status": Disabled})
}

// 激活
func (user *User) Enable(id int) (int64, error) {
	return user.Update(id, CommonMap{"status": Enabled})
}

// 验证用户名和密码
func (user *User) Match(username, password string) bool {
	err := Db.Where("(name = ? OR email = ?) AND status = ?", username, username, Enabled).First(user).Error
	if err != nil {
		return false
	}
	hashPassword := user.encryptPassword(password, user.Salt)

	return hashPassword == user.Password
}

// 获取用户详情
func (user *User) Find(id int) error {
	return Db.First(user, id).Error
}

// 用户名是否存在
func (user *User) UsernameExists(username string, uid int) (int64, error) {
	var count int64
	query := Db.Model(&User{}).Where("name = ?", username)
	if uid > 0 {
		query = query.Where("id != ?", uid)
	}
	err := query.Count(&count).Error
	return count, err
}

// 邮箱地址是否存在
func (user *User) EmailExists(email string, uid int) (int64, error) {
	var count int64
	query := Db.Model(&User{}).Where("email = ?", email)
	if uid > 0 {
		query = query.Where("id != ?", uid)
	}
	err := query.Count(&count).Error
	return count, err
}

func (user *User) List(params CommonMap) ([]User, error) {
	user.parsePageAndPageSize(params)
	list := make([]User, 0)
	err := Db.Order("id DESC").Limit(user.PageSize).Offset(user.pageLimitOffset()).Find(&list).Error

	return list, err
}

func (user *User) Total() (int64, error) {
	var count int64
	err := Db.Model(&User{}).Count(&count).Error
	return count, err
}

// 密码加密
func (user *User) encryptPassword(password, salt string) string {
	return utils.Md5(password + salt)
}

// 生成密码盐值
func (user *User) generateSalt() string {
	return utils.RandString(PasswordSaltLength)
}
