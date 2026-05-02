package Storage

import (
	"fmt"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// CreateLVM 流程：PV -> VG -> LV
func CreateLVM(c *gin.Context) {
	var req struct {
		Devices []string `json:"devices"` // 参与的物理磁盘，如 ["/dev/sdb", "/dev/sdc"]
		VGName  string   `json:"vg_name"` // 卷组名，如 "nas_data"
		LVName  string   `json:"lv_name"` // 逻辑卷名，如 "storage"
		Size    string   `json:"size"`    // 大小，如 "100G" 或 "100%FREE"
	}
	c.ShouldBindJSON(&req)

	// 1. 创建物理卷 (PV)
	for _, dev := range req.Devices {
		if err := exec.Command("pvcreate", "-f", dev).Run(); err != nil {
			c.JSON(500, gin.H{"msg": fmt.Sprintf("PV创建失败: %s", dev)})
			return
		}
	}

	// 2. 创建卷组 (VG)
	// vgcreate nas_data /dev/sdb /dev/sdc
	vgArgs := append([]string{req.VGName}, req.Devices...)
	if err := exec.Command("vgcreate", vgArgs...).Run(); err != nil {
		c.JSON(500, gin.H{"msg": "VG创建失败"})
		return
	}

	// 3. 创建逻辑卷 (LV)
	// lvcreate -L 100G -n storage nas_data  或者 -l 100%FREE
	sizeFlag := "-L"
	if req.Size == "100%FREE" {
		sizeFlag = "-l"
	}
	if err := exec.Command("lvcreate", sizeFlag, req.Size, "-n", req.LVName, req.VGName).Run(); err != nil {
		c.JSON(500, gin.H{"msg": "LV创建失败"})
		return
	}

	c.JSON(200, gin.H{"status": 0, "msg": "LVM 创建成功，路径为: /dev/" + req.VGName + "/" + req.LVName})
}

// GetLVMStatus 查看当前的 LVM 拓扑
func GetLVMStatus(c *gin.Context) {
	// vgs --json 获取卷组信息
	vgOut, _ := exec.Command("vgs", "--reportformat", "json").Output()
	// lvs --json 获取逻辑卷信息
	lvOut, _ := exec.Command("lvs", "--reportformat", "json").Output()
	// pvs --json 获取物理卷信息
	pvOut, _ := exec.Command("pvs", "--reportformat", "json").Output()

	c.JSON(200, gin.H{
		"vgs": string(vgOut),
		"lvs": string(lvOut),
		"pvs": string(pvOut),
	})
}
