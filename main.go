package main

import (
	Badger "LiteNAS/badger"
	orm "LiteNAS/database"
	"LiteNAS/router"
	"LiteNAS/utils"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	CurrentPath, _ := utils.GetCurrentPath()

	confYaml, err := utils.CheckConfig(CurrentPath)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Duration(10) * time.Second)
		os.Exit(0)
	}
	pwd := utils.MD5(strings.Join([]string{confYaml.AdminPWD, confYaml.SECRET_KEY}, ""))
	orm.InitDB(pwd)
	// gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)
	defer orm.Eloquent.Close()
	defer Badger.BadgerDB.Close()
	app := router.InitRouter(confYaml.SECRET_KEY, CurrentPath, confYaml.FormMemory)

	srv := &http.Server{
		Addr:    strings.Join([]string{":", confYaml.Port}, ""),
		Handler: app,
	}
	log.Printf("listen port %s\n", srv.Addr)
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
