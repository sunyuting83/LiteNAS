package File

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func Download(c *gin.Context) {
	// 1. 获取中间件校验后的物理路径
	realPath := c.GetString("validated_path")

	// 2. 检查是否是目录（禁止直接下载目录，目录应该先压缩）
	info, err := os.Stat(realPath)
	if err != nil || info.IsDir() {
		c.JSON(http.StatusForbidden, gin.H{"status": 1, "msg": "无法下载：路径不存在或为目录"})
		return
	}

	// 3. 设置下载头（让浏览器弹出保存框，而不是直接打开预览）
	fileName := filepath.Base(realPath)
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")

	// 4. 发送文件
	c.File(realPath)
}
