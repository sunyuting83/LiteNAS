package Storage

import (
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckLVMHealth 快速检查逻辑卷状态
func CheckLVMHealth(c *gin.Context) {
	// 使用 lvs 命令检查 Attr 字段，看是否有 'a' (active)
	out, err := exec.Command("lvs", "--noheadings", "-o", "lv_attr,lv_name").Output()
	if err != nil {
		c.JSON(500, gin.H{"msg": "无法读取 LVM 状态"})
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	healthData := make(map[string]string)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			status := "Healthy"
			// LVM 状态位首字母 'a' 代表 Active
			if !strings.HasPrefix(fields[0], "a") {
				status = "Degraded/Inactive"
			}
			healthData[fields[1]] = status
		}
	}

	c.JSON(200, gin.H{"status": 0, "data": healthData})
}
