package Storage

import (
	"os/exec"

	"github.com/gin-gonic/gin"
)

// GetDiskSmartInfo 获取指定物理盘的 SMART 信息
func GetDiskSmartInfo(c *gin.Context) {
	device := c.Query("device") // 例如 /dev/sda
	if device == "" {
		c.JSON(400, gin.H{"msg": "请提供设备路径"})
		return
	}

	// smartctl -a -j /dev/sda (-j 表示输出 JSON)
	out, err := exec.Command("smartctl", "-a", "-j", device).Output()
	if err != nil {
		// 注意：某些盘不支持 SMART 也会报错，所以我们要返回部分数据
		c.JSON(200, gin.H{"status": 1, "msg": "该设备可能不支持SMART", "raw": string(out)})
		return
	}

	// 直接转发 smartctl 的原汁原味 JSON 给前端，前端有很多成熟的图表库可以渲染它
	c.Data(200, "application/json", out)
}
