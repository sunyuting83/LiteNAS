package System

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func CollectStats() interface{} {
	// 获取 CPU 使用率
	percent, _ := cpu.Percent(0, false)
	cpuUsage := 0.0
	if len(percent) > 0 {
		cpuUsage = percent[0]
	}

	// 获取内存信息
	v, _ := mem.VirtualMemory()

	// 获取前 10 个高负载进程
	procs, _ := process.Processes()
	var procList []map[string]interface{}
	for i, p := range procs {
		if i > 10 {
			break
		} // 限制数量防止数据包过大
		name, _ := p.Name()
		cp, _ := p.CPUPercent()
		mp, _ := p.MemoryPercent()
		procList = append(procList, map[string]interface{}{
			"pid":  p.Pid,
			"name": name,
			"cpu":  cp,
			"mem":  mp,
		})
	}

	return map[string]interface{}{
		"cpu":     cpuUsage,
		"mem":     v.UsedPercent,
		"procs":   procList,
		"updated": time.Now().Format("15:04:05"),
	}
}
