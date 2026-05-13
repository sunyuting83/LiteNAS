package database

import (
	"time"
)

// Manager 用户管理表
type Manager struct {
	ID       uint   `gorm:"primaryKey"`
	UserName string `gorm:"uniqueIndex;size:64"` // 用户名唯一
	Password string `gorm:"size:128"`            // 存储哈希后的密码
	// UserPath 建议存储更详细的 JSON 配置，包含多个挂载点的权限
	// 格式示例: [{"path": "/mnt/disk1/data", "label": "我的数据", "read": true, "write": true}]
	UserPath  string `gorm:"type:text"`
	Status    int    `gorm:"default:0"`     // 0:正常, 1:禁用
	IsAdmin   bool   `gorm:"default:false"` // 是否为管理员
	LastLogin int64  `gorm:"index"`         // 最后登录时间
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
}

func (manager *Manager) Insert() (err error) {
	DB.Create(&manager)
	return nil
}

// CheckAdminLogin 供 controller 调用
func CheckAdminLogin(username, password string) (*Manager, error) {
	var user Manager
	// 查询用户且状态正常
	err := DB.Where("user_name = ? AND password = ? AND status = 0", username, password).First(&user).Error
	if err != nil {
		return nil, err
	}
	// 更新最后登录时间
	DB.Model(&user).Update("last_login", time.Now().UnixMilli())
	return &user, nil
}

func CheckUserName(username string) (manager *Manager, err error) {
	if err = DB.First(&manager, "user_name = ? ", username).Error; err != nil {
		return
	}
	return
}

func CheckUserID(id uint) (manager *Manager, err error) {
	if err = DB.First(&manager, "id = ? ", id).Error; err != nil {
		return
	}
	return
}

// Get Count
func (manager *Manager) GetCount() (count int64, err error) {
	if err = DB.Model(&manager).Count(&count).Error; err != nil {
		return
	}
	return
}

// Check ID
func CheckID(id int64) (manager *Manager, err error) {
	if err = DB.First(&manager, "id = ?", id).Error; err != nil {
		return
	}
	return
}

// Delete Admin
func (manager *Manager) DeleteOne(id int64) {
	// time.Sleep(time.Duration(100) * time.Millisecond)
	DB.Where("id = ?", id).Delete(&manager)
}

// Admin List
func GetAdminList(page, Limit int) (manages *[]Manager, err error) {
	p := makePage(page, Limit)
	if err = DB.
		Select("id, user_name, new_status, created_at").
		Order("id desc").
		Limit(Limit).Offset(p).
		Find(&manages).Error; err != nil {
		return
	}
	return
}

// Admin List
func GetHasUsersID(manager_id uint) (manages *Manager, err error) {
	if err = DB.
		Where(&Manager{ID: manager_id}).Find(&manages).Error; err != nil {
		return
	}
	return
}

// Reset Password
func (manager *Manager) ResetPassword(username string) (manage Manager, err error) {
	// time.Sleep(time.Duration(100) * time.Millisecond)
	if err = DB.First(&manage, "user_name = ?", username).Error; err != nil {
		return
	}
	// fmt.Println(manager)
	if err = DB.Model(&manage).Updates(&manager).Error; err != nil {
		return
	}
	return
}

// Reset Password
func (manager *Manager) UpStatusAdmin(status int) {
	DB.Model(&manager).Update("new_status", status)
}

// makePage make page
func makePage(p, Limit int) int {
	p = max(p-1, 0)
	page := p * Limit
	return page
}
