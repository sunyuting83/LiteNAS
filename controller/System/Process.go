package System

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/process"
)

func KillProcess(c *gin.Context) {
	var req struct {
		Pid int32 `json:"pid"`
	}
	c.ShouldBindJSON(&req)

	// 1. 安全检查：禁止杀掉自己
	if req.Pid == int32(os.Getpid()) {
		c.JSON(403, gin.H{"msg": "禁止自杀行为！"})
		return
	}

	// 2. 安全检查：禁止杀掉系统关键进程 (PID 小于 100)
	if req.Pid < 100 {
		c.JSON(403, gin.H{"msg": "系统核心进程，拒绝操作"})
		return
	}

	p, err := process.NewProcess(req.Pid)
	if err != nil {
		c.JSON(404, gin.H{"msg": "进程不存在"})
		return
	}

	// 3. 安全检查：检查进程名
	name, _ := p.Name()
	critical := []string{"sshd", "systemd", "init", "dbus", "LiteNAS"}
	for _, cName := range critical {
		if name == cName {
			c.JSON(403, gin.H{"msg": "关键服务禁止停止"})
			return
		}
	}

	// 执行删除
	if err := p.Kill(); err != nil {
		c.JSON(500, gin.H{"msg": "Kill 失败"})
		return
	}

	c.JSON(200, gin.H{"status": 1})
}
