package middleware

import (
	"LiteNAS/database"
	sqlDB "LiteNAS/database"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// GlobalDataRoot 系统强制的数据根目录，禁止任何 API 超出此范围
const GlobalDataRoot = "/mnt"

// UserPermission 对应数据库中 Manager 表的 UserPath JSON 结构
type UserPermission struct {
	Path  string `json:"path"`
	Label string `json:"label"`
	Read  bool   `json:"read"`
	Write bool   `json:"write"`
}

// FilePolicy 文件操作权限校验中间件
func FilePolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取由 Auth 中间件存入的 UserID
		uidRaw, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"status": 1, "message": "未登录"})
			c.Abort()
			return
		}
		uid := uidRaw.(uint)

		// 2. 获取请求路径
		reqPath := c.DefaultQuery("path", c.PostForm("path"))
		if reqPath == "" {
			// 如果某些接口不需要路径参数也可以通过，可放行
			c.Next()
			return
		}

		// 3. 深度解析：评估软链接并转为绝对路径
		// filepath.Join 会处理掉相对路径，EvalSymlinks 会还原软链接的真实物理路径
		absPath := filepath.Join(GlobalDataRoot, reqPath)
		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			// 如果路径不存在（比如用户正在创建新文件夹），EvalSymlinks 会报错
			// 此时我们回退到使用 Clean 保证路径格式正确
			realPath = filepath.Clean(absPath)
		}

		// 4. 强制校验：是否超出了系统允许的数据根目录（防止访问 /etc 等）
		if !strings.HasPrefix(realPath, GlobalDataRoot) {
			c.JSON(http.StatusForbidden, gin.H{"status": 1, "message": "非法操作：严禁越权访问系统目录"})
			c.Abort()
			return
		}

		// 5. 获取用户数据库权限
		var user *database.Manager
		if err := sqlDB.DB.First(&user, "id = ?", uid).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"status": 1, "message": "用户数据异常"})
			c.Abort()
			return
		}

		// 6. 权限比对逻辑
		authorized := false
		isWrite := isWriteMethod(c.Request.Method)

		// 管理员特权：只要在 GlobalDataRoot 范围内就放行
		if user.IsAdmin {
			authorized = true
		} else {
			// 普通用户解析 UserPath JSON
			var perms []UserPermission
			if err := json.Unmarshal([]byte(user.UserPath), &perms); err == nil {
				for _, p := range perms {
					// 将授权路径也转为清洁的绝对路径
					authPath := filepath.Clean(p.Path)
					// 检查请求路径是否在授权路径之下
					if strings.HasPrefix(realPath, authPath) {
						if isWrite && !p.Write {
							continue // 有读取权但无写入权，继续匹配其他可能的授权
						}
						if !isWrite && !p.Read {
							continue
						}
						authorized = true
						break
					}
				}
			}
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"status": 1, "message": "无访问权限或目录未授权"})
			c.Abort()
			return
		}

		// 7. 将校验后的“真实物理路径”存入 Context，Controller 直接使用 c.GetString("validated_path")
		c.Set("validated_path", realPath)
		c.Next()
	}
}

// isWriteMethod 辅助函数：判断是否为写操作请求
func isWriteMethod(method string) bool {
	m := strings.ToUpper(method)
	// POST(创建/修改), PUT(覆盖), DELETE(删除), PATCH(局部修改)
	return m == "POST" || m == "PUT" || m == "DELETE" || m == "PATCH"
}
