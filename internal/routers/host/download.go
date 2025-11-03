package host

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/models"
)

// DownloadAgent 下载 gocron-node 二进制文件
func DownloadAgent(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Missing token parameter")
		return
	}

	// 验证 token
	settingModel := new(models.Setting)
	savedToken := settingModel.GetAgentRegisterToken()
	if savedToken == "" || token != savedToken {
		c.String(http.StatusForbidden, "Invalid token")
		return
	}

	os := c.DefaultQuery("os", "linux")
	arch := c.DefaultQuery("arch", "amd64")

	// 查找二进制文件
	binaryPath := findAgentBinary(os, arch)
	if binaryPath == "" {
		c.String(http.StatusNotFound, fmt.Sprintf("Agent binary not found for %s-%s", os, arch))
		return
	}

	// 检查文件是否存在
	if !utils.FileExist(binaryPath) {
		c.String(http.StatusNotFound, "Agent binary file not found")
		return
	}

	// 设置下载文件名
	filename := fmt.Sprintf("gocron-node-%s-%s.tar.gz", os, arch)
	if os == "windows" {
		filename = fmt.Sprintf("gocron-node-%s-%s.zip", os, arch)
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.File(binaryPath)
}

// findAgentBinary 查找 agent 二进制文件
func findAgentBinary(os, arch string) string {
	// 在 gocron-node-package 目录中查找
	baseDir := "gocron-node-package"
	
	var pattern string
	if os == "windows" {
		pattern = fmt.Sprintf("gocron-node-*-%s-%s.zip", os, arch)
	} else {
		pattern = fmt.Sprintf("gocron-node-*-%s-%s.tar.gz", os, arch)
	}

	matches, err := filepath.Glob(filepath.Join(baseDir, pattern))
	if err != nil || len(matches) == 0 {
		return ""
	}

	return matches[0]
}
