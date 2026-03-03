package router

import (
	Users "LiteNAS/controller/Users"

	"github.com/gin-gonic/gin"
)

// InitRouter make router
func InitRouter(secretKey string, currentPath string, formMemory int64) *gin.Engine {
	r := gin.Default()

	// 设置上传限制
	r.MaxMultipartMemory = formMemory << 20 // MB 转字节

	// 开放登录接口
	public := r.Group("/api")
	{
		public.POST("/login", Users.Sgin(secretKey)) // 登录接口放在这里
	}
	/*
				// 2. 鉴权组：必须有 Token 才能访问
		    private := r.Group("/api")
		    private.Use(middleware.AuthMiddleware(secretKey)) // 挂载中间件
		    {
		        private.GET("/file/list", controller.ListDir)
		        // ... 其他需要权限的接口
		    }
	*/

	return r
}
