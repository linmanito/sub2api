package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadDir string
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler() *UploadHandler {
	// 默认上传目录为 data/uploads
	uploadDir := os.Getenv("DATA_DIR")
	if uploadDir == "" {
		uploadDir = "./data"
	}
	uploadDir = filepath.Join(uploadDir, "uploads")

	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		// 日志记录但不阻止启动
		fmt.Printf("Warning: failed to create upload directory %s: %v\n", uploadDir, err)
	}

	return &UploadHandler{uploadDir: uploadDir}
}

// UploadImage 上传图片
// POST /api/v1/admin/upload/image
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded: "+err.Error())
		return
	}
	defer file.Close()

	// 验证文件类型
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		response.BadRequest(c, "Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed")
		return
	}

	// 验证文件大小（最大 5MB）
	const maxSize = 5 * 1024 * 1024
	if header.Size > maxSize {
		response.BadRequest(c, "File too large. Maximum size is 5MB")
		return
	}

	// 生成唯一文件名
	ext := getImageExtension(contentType)
	filename := generateFilename(ext)

	// 确保上传目录存在
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		response.InternalError(c, "Failed to create upload directory")
		return
	}

	// 保存文件
	filePath := filepath.Join(h.uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		response.InternalError(c, "Failed to create file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // 清理失败的文件
		response.InternalError(c, "Failed to save file")
		return
	}

	// 返回文件路径（相对于 uploads 目录）
	response.Success(c, gin.H{
		"filename": filename,
		"path":     "/uploads/" + filename,
	})
}

// isValidImageType 验证是否为有效的图片类型
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

// getImageExtension 根据内容类型获取文件扩展名
func getImageExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}

// generateFilename 生成随机文件名
func generateFilename(ext string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes) + ext
}

// ServeUploads 提供静态文件服务的中间件
// 用于 /uploads/* 路由
func ServeUploads() gin.HandlerFunc {
	uploadDir := os.Getenv("DATA_DIR")
	if uploadDir == "" {
		uploadDir = "./data"
	}
	uploadDir = filepath.Join(uploadDir, "uploads")

	return func(c *gin.Context) {
		// 获取请求的文件名
		filename := strings.TrimPrefix(c.Request.URL.Path, "/uploads/")
		if filename == "" {
			response.NotFound(c, "File not found")
			c.Abort()
			return
		}

		// 安全检查：防止路径遍历攻击
		filename = filepath.Base(filename)
		filePath := filepath.Join(uploadDir, filename)

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			response.NotFound(c, "File not found")
			c.Abort()
			return
		}

		// 设置缓存头
		c.Header("Cache-Control", "public, max-age=31536000")

		c.File(filePath)
	}
}
