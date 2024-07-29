package database

import "fmt"

type Manager struct {
	ID        uint   `gorm:"primaryKey"`
	UserName  string `gorm:"index:idx_manager_user_pass_status;index:idx_manager_username"`
	Password  string `gorm:"index:idx_manager_user_pass_status"`
	UserPath  string
	NewStatus int   `gorm:"index:idx_manager_user_pass_status"`
	CreatedAt int64 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
}

func (manager *Manager) Insert() (err error) {
	sqlDB.Create(&manager)
	return nil
}

func CheckAdminLogin(username, password string) (manager *Manager, err error) {
	if err = sqlDB.First(&manager, "user_name = ? AND new_status = ? AND password = ?", username, "0", password).Error; err != nil {
		return
	}
	return
}

func CheckUserName(username string) (manager *Manager, err error) {
	if err = sqlDB.First(&manager, "user_name = ? ", username).Error; err != nil {
		return
	}
	return
}

func CheckUserID(id uint) (manager *Manager, err error) {
	if err = sqlDB.First(&manager, "id = ? ", id).Error; err != nil {
		return
	}
	return
}

// Get Count
func (manager *Manager) GetCount() (count int64, err error) {
	if err = sqlDB.Model(&manager).Count(&count).Error; err != nil {
		return
	}
	return
}

// Check ID
func CheckID(id int64) (manager *Manager, err error) {
	if err = sqlDB.First(&manager, "id = ?", id).Error; err != nil {
		return
	}
	return
}

// Delete Admin
func (manager *Manager) DeleteOne(id int64) {
	// time.Sleep(time.Duration(100) * time.Millisecond)
	sqlDB.Where("id = ?", id).Delete(&manager)
}

// Admin List
func GetAdminList(page, Limit int) (manages *[]Manager, err error) {
	p := makePage(page, Limit)
	if err = sqlDB.
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
	if err = sqlDB.
		Where(&Manager{ID: manager_id}).Find(&manages).Error; err != nil {
		return
	}
	return
}

// Reset Password
func (manager *Manager) ResetPassword(username string) (manage Manager, err error) {
	// time.Sleep(time.Duration(100) * time.Millisecond)
	if err = sqlDB.First(&manage, "user_name = ?", username).Error; err != nil {
		return
	}
	fmt.Println(manager)
	if err = sqlDB.Model(&manage).Updates(&manager).Error; err != nil {
		return
	}
	return
}

// Reset Password
func (manager *Manager) UpStatusAdmin(status int) {
	sqlDB.Model(&manager).Update("new_status", status)
}

// makePage make page
func makePage(p, Limit int) int {
	p = p - 1
	if p <= 0 {
		p = 0
	}
	page := p * Limit
	return page
}
