package Storage

import (
	"encoding/json"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// LsblkOutput lsblk JSON 输出的最外层
type LsblkOutput struct {
	Blockdevices []DiskDevice `json:"blockdevices"`
}

// DiskDevice 对应物理磁盘 (如 /dev/sda)
type DiskDevice struct {
	Name     string      `json:"name"`     // 设备名
	Model    string      `json:"model"`    // 型号
	Size     string      `json:"size"`     // 总大小
	PTType   string      `json:"pttype"`   // 分区表类型 (gpt/dos)
	PTUUID   string      `json:"ptuuid"`   // 磁盘标识符
	Type     string      `json:"type"`     // 类型 (disk)
	Children []Partition `json:"children"` // 分区列表
}

// Partition 对应分区 (如 /dev/sda1)
type Partition struct {
	Name         string `json:"name"`         // 分区名
	Size         string `json:"size"`         // 分区大小
	FSType       string `json:"fstype"`       // 类型 (ext4/ntfs)
	UUID         string `json:"uuid"`         // UUID
	PartTypeName string `json:"parttypename"` // 分区类型名 (Linux 文件系统)
	Mountpoint   string `json:"mountpoint"`   // 挂载点
}

// GetDiskLayout 获取类似 fdisk -l + blkid 的组合结果
func GetDiskLayout(c *gin.Context) {
	// 核心命令：lsblk -J (JSON格式) -p (完整路径) -o (指定字段)
	// 字段说明：name(路径), size(大小), model(型号), pttype(标签类型), ptuuid(标识符), fstype(文件系统), uuid(UUID), parttypename(类型名)
	args := []string{
		"-J", "-p", "-b", // -b 是为了拿到准确的 Byte 数，如果你想直接要 GiB 可以去掉 -b
		"-o", "NAME,SIZE,MODEL,PTTYPE,PTUUID,FSTYPE,UUID,PARTTYPENAME,TYPE,MOUNTPOINT",
	}

	cmd := exec.Command("lsblk", args...)
	out, err := cmd.Output()
	if err != nil {
		c.JSON(500, gin.H{"msg": "Failed to execute lsblk", "error": err.Error()})
		return
	}

	var raw LsblkOutput
	if err := json.Unmarshal(out, &raw); err != nil {
		c.JSON(500, gin.H{"msg": "Failed to parse disk data"})
		return
	}

	// 过滤：只保留物理磁盘 (disk)，过滤掉只读的 loop 或 rom
	var filtered []DiskDevice
	for _, dev := range raw.Blockdevices {
		if dev.Type == "disk" {
			filtered = append(filtered, dev)
		}
	}

	c.JSON(200, gin.H{
		"status": 0,
		"data":   filtered,
	})
}
