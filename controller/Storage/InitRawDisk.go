package Storage

import (
	"os/exec"

	"github.com/gin-gonic/gin"
)

// InitRawDisk 处理完全空白的硬盘
func InitRawDisk(c *gin.Context) {
	var req struct {
		Device string `json:"device"` // 例如 /dev/sdb (注意是整个磁盘，不是 sdb1)
	}
	c.ShouldBindJSON(&req)

	// 安全校验：严禁操作系统盘
	if req.Device == "/dev/sda" {
		c.JSON(403, gin.H{"msg": "拒绝初始化系统主盘"})
		return
	}

	// 1. 创建 GPT 分区表 (适合大容量硬盘)
	// parted -s /dev/sdb mklabel gpt
	cmdLabel := exec.Command("parted", "-s", req.Device, "mklabel", "gpt")
	if err := cmdLabel.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "创建分区表失败: " + err.Error()})
		return
	}

	// 2. 将整个硬盘划分为一个主分区
	// parted -s /dev/sdb mkpart primary ext4 0% 100%
	cmdPart := exec.Command("parted", "-s", req.Device, "mkpart", "primary", "ext4", "0%", "100%")
	if err := cmdPart.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "创建分区失败"})
		return
	}

	c.JSON(200, gin.H{"status": 0, "msg": "分区创建成功，请开始格式化"})
}
