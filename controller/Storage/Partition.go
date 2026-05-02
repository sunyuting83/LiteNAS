package Storage

import (
	"os/exec"

	"github.com/gin-gonic/gin"
)

// FormatDisk 格式化分区 (谨慎使用！)
func FormatDisk(c *gin.Context) {
	var req struct {
		Device string `json:"device"` // /dev/sdb1
		FS     string `json:"fs"`     // ext4 或 ntfs
	}
	c.ShouldBindJSON(&req)

	// 这里可以加一个白名单，禁止格式化 sda (通常是系统盘)
	if req.Device == "/dev/sda" || req.Device == "/dev/sdb5" {
		c.JSON(403, gin.H{"msg": "检测到系统关键分区，拒绝格式化"})
		return
	}

	// 执行格式化 mkfs.ext4 /dev/sdb1
	fsCmd := "mkfs." + req.FS
	cmd := exec.Command(fsCmd, req.Device)
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "格式化失败"})
		return
	}
	c.JSON(200, gin.H{"status": 0, "msg": "格式化成功"})
}
