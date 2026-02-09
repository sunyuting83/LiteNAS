package database

import (
	"LiteNAS/utils"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB    *gorm.DB
	sqlDB *sql.DB // 内部保存 sql.DB 句柄用于关闭
)

// InitDB 初始化数据库
func InitDB(defaultAdminPwd string) {
	// 1. 使用 filepath 替代 strings.Join 保证跨平台路径正确
	currentPath, _ := utils.GetCurrentPath()
	dbDir := filepath.Join(currentPath, "db")

	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		_ = os.MkdirAll(dbDir, 0755)
	}

	dbFile := filepath.Join(dbDir, "nas_core.db")

	// 2. 开启 GORM 配置优化
	var err error
	DB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 生产环境建议关闭详细日志
	})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}

	// 3. 配置连接池
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 4. 自动迁移
	_ = DB.AutoMigrate(&Manager{})

	// 5. 初始化默认管理员
	initAdmin(defaultAdminPwd)
}

func initAdmin(pwd string) {
	var count int64
	DB.Model(&Manager{}).Count(&count)
	if count == 0 {
		// 初始权限配置 (JSON 格式)
		// 预留给主程序逻辑层解析
		defaultConfig := `[{"path":"/mnt/default","label":"根目录","read":true,"write":true}]`

		admin := Manager{
			UserName: "admin",
			Password: pwd, // 注意：这里的 pwd 应在传入前由 logic 层完成哈希
			Status:   0,
			UserPath: defaultConfig,
			IsAdmin:  true,
		}
		DB.Create(&admin)
	}
}

// CloseDB 供 main 函数 defer 调用
func CloseDB() {
	if sqlDB != nil {
		sqlDB.Close()
	}
}
