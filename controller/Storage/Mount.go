package Storage

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// MountPartition 挂载分区
func MountPartition(c *gin.Context) {
	var req struct {
		Device string `json:"device"` // /dev/sda1
		Path   string `json:"path"`   // sda1 (自定义名称)
	}
	c.ShouldBindJSON(&req)

	// 1. 统一挂载前缀
	mountPath := fmt.Sprintf("/mnt/%s", req.Path)

	// 2. 检查目录是否存在，不存在则创建
	if _, err := os.Stat(mountPath); os.IsNotExist(err) {
		os.MkdirAll(mountPath, 0755)
	}

	// 3. 执行挂载 (sudo mount /dev/sda1 /mnt/sda1)
	cmd := exec.Command("mount", req.Device, mountPath)
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "挂载失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": 0, "msg": "挂载成功", "at": mountPath})
}

// UmountPartition 卸载分区
func UmountPartition(c *gin.Context) {
	var req struct {
		Path string `json:"path"` // /mnt/sda1
	}
	c.ShouldBindJSON(&req)

	// 安全检查：严禁卸载系统根目录
	if req.Path == "/" || req.Path == "/boot" {
		c.JSON(403, gin.H{"msg": "禁止卸载系统核心分区"})
		return
	}

	cmd := exec.Command("umount", "-l", req.Path) // -l 懒卸载，防止设备忙
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "卸载失败"})
		return
	}
	c.JSON(200, gin.H{"status": 0, "msg": "已安全卸载"})
}
