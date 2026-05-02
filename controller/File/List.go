package File

import (
	"LiteNAS/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// FileItem 统一的输出结构
type FileItem struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    string `json:"size"`
	Ext     string `json:"ext"`
	ModTime string `json:"mod_time"`
}

// GetList 读取目录接口
func GetList(c *gin.Context) {
	// 1. 获取中间件校验后的物理路径 (例如: /mnt/sda1/videos)
	realPath := c.GetString("validated_path")

	// 系统根目录常量，用于计算相对路径
	const dataRoot = "/mnt"

	// 2. 读取目录内容
	entries, err := os.ReadDir(realPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 1, "msg": "读取目录失败"})
		return
	}

	var dirs []FileItem
	var files []FileItem

	for _, entry := range entries {
		name := entry.Name()
		if name[0] == '.' {
			continue
		} // 过滤隐藏文件

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 计算相对路径（去掉 /mnt 前缀），方便前端展示和下次请求
		relPath, _ := filepath.Rel(dataRoot, filepath.Join(realPath, name))

		item := FileItem{
			Name:    name,
			Path:    relPath,
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		}

		if entry.IsDir() {
			item.Ext = "dir"
			item.Size = "-"
			dirs = append(dirs, item)
		} else {
			item.Ext = filepath.Ext(name)
			item.Size = utils.FormatFileSize(info.Size())
			files = append(files, item)
		}
	}

	// 合并：文件夹排在前面
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"data":   append(dirs, files...),
	})
}
