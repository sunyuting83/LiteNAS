package File

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Delete(c *gin.Context) {
	realPath := c.GetString("validated_path")

	// 1. 防止误删根目录 (哪怕是 Admin 也不行)
	if realPath == "/mnt" || realPath == "/mnt/" {
		c.JSON(http.StatusForbidden, gin.H{"status": 1, "msg": "禁止删除数据根目录"})
		return
	}

	// 2. 执行删除
	// os.RemoveAll 可以删除文件或非空目录
	err := os.RemoveAll(realPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 1, "msg": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "删除成功"})
}
