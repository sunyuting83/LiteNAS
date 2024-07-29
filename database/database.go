package database

import (
	"LiteNAS/utils"
	"database/sql"
	"os"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Eloquent *sql.DB
	sqlDB    *gorm.DB
)

// InitDB init db
func InitDB(pwd string) {
	CurrentPath, _ := utils.GetCurrentPath()
	dbPath := strings.Join([]string{CurrentPath, "db"}, "/")
	if !utils.IsExist(dbPath) {
		os.MkdirAll(dbPath, 0755)
	}
	dbName := strings.Join([]string{"database", "sqlite3"}, ".")
	dbFile := strings.Join([]string{dbPath, dbName}, "/")
	sqlDB, _ = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

	Eloquent, _ = sqlDB.DB()
	Eloquent.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	Eloquent.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	Eloquent.SetConnMaxLifetime(time.Hour)
	sqlDB.AutoMigrate(
		&Manager{},
	)

	var (
		manager *Manager
	)

	if m := sqlDB.First(&manager); m.Error != nil {
		if m.Error.Error() == "record not found" {
			u := Manager{
				UserName:  "admin",
				Password:  pwd,
				NewStatus: 0,
			}
			sqlDB.Create(&u)
		}
	}
}
