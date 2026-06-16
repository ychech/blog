package handler

import (
	"blog/config"
	"blog/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器，目前支持图片上传与静态访问。
type UploadHandler struct {
	uploadPath string
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler() *UploadHandler {
	cfg := config.C.App
	// 确保上传目录存在
	_ = os.MkdirAll(cfg.UploadPath, 0755)
	return &UploadHandler{uploadPath: cfg.UploadPath}
}

// UploadImage 上传图片
// 限制：仅允许 jpg/png/gif/webp，最大由 MaxUploadSize 控制
// @Summary 上传图片
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} utils.Response{data=object{url=string,filename=string,size=int}}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /uploads [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	cfg := config.C.App
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxUploadSize*1024*1024)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > cfg.MaxUploadSize*1024*1024 {
		utils.BadRequest(c, fmt.Sprintf("文件大小超过限制 %d MB", cfg.MaxUploadSize))
		return
	}

	// 检查文件后缀
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		utils.BadRequest(c, "仅支持 jpg/png/gif/webp 格式")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(h.uploadPath, filename)

	if err := c.SaveUploadedFile(header, savePath); err != nil {
		utils.InternalError(c, "保存文件失败: "+err.Error())
		return
	}

	// 返回可访问 URL
	fileURL := fmt.Sprintf("/uploads/%s", filename)
	utils.Success(c, gin.H{
		"url":      fileURL,
		"filename": filename,
		"size":     header.Size,
	})
}
