package utils

import "fmt"

// FormatFileSize 高性能文件大小格式化
func FormatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2fGB", float64(size)/(1024*1024*1024))
}
