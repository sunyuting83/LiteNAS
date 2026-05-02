package router

import (
	"LiteNAS/controller/File"
	"LiteNAS/controller/Storage"
	"LiteNAS/controller/Users"
	"LiteNAS/middleware"

	"github.com/gin-gonic/gin"
)

// InitRouter make router
func InitRouter(secretKey string, currentPath string, formMemory int64) *gin.Engine {
	r := gin.Default()

	// 设置上传限制
	r.MaxMultipartMemory = formMemory << 20 // MB 转字节

	middleware.InitWS() // 初始化 WS 管理器

	// 开放登录接口
	public := r.Group("/api")
	{
		public.POST("/login", Users.Sgin(secretKey)) // 登录接口放在这里

	}
	private := r.Group("/api")
	private.Use(middleware.AuthMiddleware(secretKey))
	{
		// 文件管理相关的路由，额外挂载 FilePolicy 中间件
		fileApi := private.Group("/file")
		fileApi.Use(middleware.FilePolicy())
		{
			// 对应 controller/File/List.go 中的 GetList 函数
			fileApi.GET("/list", File.GetList)
			fileApi.GET("/download", File.Download)
			fileApi.POST("/upload", File.Upload)
			fileApi.POST("/delete", File.Delete)
		}

		systemApi := private.Group("/system") // private 组已经使用了 AuthMiddleware
		{
			systemApi.GET("/monitor", func(c *gin.Context) {
				// 执行到这里，说明 AuthMiddleware 已经验证 Token 通过了
				// 我们只需要在这里完成协议升级
				conn, err := middleware.WSUpgrader.Upgrade(c.Writer, c.Request)
				if err != nil {
					return
				}
				go conn.ReadLoop()
			})
		}

		storageApi := private.Group("/storage")
		{
			storageApi.GET("/list", Storage.GetDiskLayout)      // 查看磁盘状态
			storageApi.POST("/mount", Storage.MountPartition)   // 挂载
			storageApi.POST("/umount", Storage.UmountPartition) // 卸载
			storageApi.POST("/format", Storage.FormatDisk)      // 格式化
			storageApi.POST("/init", Storage.InitRawDisk)       // 初始化新盘
		}

	}

	return r
}
