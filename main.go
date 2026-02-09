package main

import (
	Badger "LiteNAS/badger"
	"LiteNAS/database"
	"LiteNAS/router"
	"LiteNAS/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 获取程序运行目录
	currentPath, err := os.Getwd() // 或者继续用你的 utils.GetCurrentPath
	if err != nil {
		log.Fatalf("获取目录失败: %v", err)
	}

	// 2. 加载/初始化配置
	conf, err := utils.CheckConfig(currentPath)
	if err != nil {
		log.Printf("配置文件错误: %v\n程序将在10秒后退出", err)
		time.Sleep(10 * time.Second)
		return
	}

	// 3. 初始化数据库
	// 密码哈希: 建议在 logic 层做，这里为了兼容你之前的逻辑
	adminHash := utils.MD5(fmt.Sprintf("%s%s", conf.AdminPWD, conf.SECRET_KEY))
	database.InitDB(adminHash)

	// 4. 延迟关闭数据库资源 (修复 Close 问题)
	defer database.CloseDB()
	defer Badger.BadgerDB.Close()

	// 5. 初始化路由
	gin.SetMode(gin.DebugMode)
	app := router.InitRouter(conf.SECRET_KEY, currentPath, conf.FormMemory)

	srv := &http.Server{
		Addr:    ":" + conf.Port,
		Handler: app,
	}

	// 6. 优雅启动
	go func() {
		log.Printf("LiteNAS Core 运行在端口 %s\n", conf.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 7. 信号监听 (支持 Interrupt 和 Terminate)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("LiteNAS 已安全退出")
}
