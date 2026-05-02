package File

import (
	"LiteNAS/middleware"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ProgressWriter 用于计算写入进度
type ProgressWriter struct {
	Total      int64
	Written    int64
	ID         string // WebSocket 的 UUID
	FileName   string
	LastUpdate int // 用于减少推送频率，每增长 1% 推送一次
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Written += int64(n)

	// 计算百分比
	percent := int(float64(pw.Written) / float64(pw.Total) * 100)

	// 频率控制：只有进度增加时才推送，避免挤爆 WebSocket 通道
	if percent > pw.LastUpdate {
		pw.LastUpdate = percent
		// 推送消息给特定用户
		msg := fmt.Sprintf(`{"filename":"%s", "percent":%d}`, pw.FileName, percent)
		middleware.Manager.SendToID(pw.ID, "upload_progress", msg)
	}
	return n, nil
}

func Upload(c *gin.Context) {
	targetDir := c.GetString("validated_path")
	wsID := c.GetHeader("X-Workspace-ID")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 1, "msg": "未接收到文件"})
		return
	}

	// 安全文件名处理
	safeFileName := filepath.Base(file.Filename)
	dst := filepath.Join(targetDir, safeFileName)

	// 打开源文件流
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"msg": "打开上传文件失败"})
		return
	}
	defer src.Close()

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(500, gin.H{"msg": "创建文件失败"})
		return
	}
	defer out.Close()

	// 核心：接入进度追踪
	// 如果前端传了 ws_id，我们才进行进度推送
	var writer io.Writer = out
	if wsID != "" {
		writer = io.MultiWriter(out, &ProgressWriter{
			Total:    file.Size,
			ID:       wsID,
			FileName: safeFileName,
		})
	}

	// 执行拷贝
	_, err = io.Copy(writer, src)
	if err != nil {
		c.JSON(500, gin.H{"msg": "上传中断"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 0, "msg": "上传成功"})
}
