package Storage

import (
	"fmt"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// RemoveLV 删除逻辑卷 (危险操作！)
func RemoveLV(c *gin.Context) {
	var req struct {
		LVPath string `json:"lv_path"` // 例如 /dev/nas_data/storage
	}
	c.ShouldBindJSON(&req)

	// 1. 必须先卸载，否则无法删除
	// 这里可以调用之前的 Umount 逻辑，或者直接执行
	exec.Command("umount", "-l", req.LVPath).Run()

	// 2. 执行 lvremove -f (强制删除)
	cmd := exec.Command("lvremove", "-f", req.LVPath)
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "删除逻辑卷失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": 0, "msg": "逻辑卷已成功删除"})
}

// ExtendLVM 扩容逻辑
func ExtendLVM(c *gin.Context) {
	var req struct {
		VGName string `json:"vg_name"` // 卷组名
		LVName string `json:"lv_name"` // 逻辑卷名
		NewDev string `json:"new_dev"` // 新加入的物理设备，如 /dev/sdd (可选)
		Size   string `json:"size"`    // 扩容多少，如 "+50G" 或 "100%FREE"
	}
	c.ShouldBindJSON(&req)

	// 1. 如果有新设备，先将其加入 VG (vgextend)
	if req.NewDev != "" {
		exec.Command("pvcreate", "-f", req.NewDev).Run()
		if err := exec.Command("vgextend", req.VGName, req.NewDev).Run(); err != nil {
			c.JSON(500, gin.H{"msg": "扩展卷组失败"})
			return
		}
	}

	// 2. 扩展逻辑卷 (lvextend)
	// -r 参数极其重要：它会在扩容后自动执行 resize2fs (ext4) 或 xfs_growfs，实现“在线扩容”
	lvPath := fmt.Sprintf("/dev/%s/%s", req.VGName, req.LVName)
	cmd := exec.Command("lvextend", "-r", "-l", req.Size, lvPath)
	if req.Size[0] == '+' || req.Size[0] == '-' { // 如果是 +50G 这种格式
		cmd = exec.Command("lvextend", "-r", "-L", req.Size, lvPath)
	}

	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"msg": "扩容失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": 0, "msg": "扩容成功，空间已即时生效"})
}
