package agent

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/routers/base"
)

const tokenExpiration = 3 * time.Hour

// GenerateToken 生成注册token
func GenerateToken(c *gin.Context) {
	token := generateRandomToken()
	expiresAt := time.Now().Add(tokenExpiration)

	agentToken := &models.AgentToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := agentToken.Create(); err != nil {
		logger.Error("创建token失败:", err)
		base.RespondError(c, i18n.T(c, "operation_failed"), err)
		return
	}

	serverURL := getServerURL(c)
	installCmdLinux := fmt.Sprintf("curl -fsSL '%s/api/agent/install.sh?token=%s' | bash", serverURL, token)

	base.RespondSuccess(c, i18n.T(c, "operation_success"), map[string]interface{}{
		"token":       token,
		"expires_at":  expiresAt,
		"install_cmd": installCmdLinux,
	})
}

// InstallScript 返回安装脚本
func InstallScript(c *gin.Context) {
	// 验证token
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Token is required")
		return
	}

	// 验证token有效性
	agentToken := &models.AgentToken{}
	if err := agentToken.FindByToken(token); err != nil {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}

	if time.Now().After(agentToken.ExpiresAt) {
		c.String(http.StatusUnauthorized, "Token expired")
		return
	}

	script := `#!/bin/bash
set -e

# 安全检查：禁止使用 root 用户运行
if [ "$(id -u)" = "0" ]; then
    echo "Error: This script should NOT be run as root for security reasons."
    echo "Please run as a regular user with sudo privileges."
    echo "Example: su - youruser -c 'curl -fsSL ... | bash'"
    exit 1
fi

# Token is embedded in the script URL, extract it here
TOKEN="` + token + `"
if [ -z "$TOKEN" ]; then
    echo "Error: Token is required"
    exit 1
fi

GOCRON_SERVER="` + getServerURL(c) + `"
INSTALL_DIR="/opt/gocron-node"
SERVICE_NAME="gocron-node"

ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    echo "This script is for Linux/macOS. For Windows, use PowerShell script."
    echo "PowerShell command:"
    echo "  iwr -useb ` + getServerURL(c) + `/api/agent/install.ps1 | iex"
    exit 1
fi

echo "Installing gocron-node for $OS-$ARCH..."

# 检测本地服务器是否有安装包
LOCAL_DOWNLOAD_URL="${GOCRON_SERVER}/api/agent/download?os=${OS}&arch=${ARCH}"
echo "Checking local server for installation package..."

# 使用 HEAD 请求检测，-w %{http_code} 获取状态码，-o /dev/null 不输出内容
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$LOCAL_DOWNLOAD_URL")

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if [ "$HTTP_CODE" = "200" ]; then
    # 本地有安装包，直接下载
    echo "✓ Local package found, downloading from local server..."
    DOWNLOAD_URL="$LOCAL_DOWNLOAD_URL"
elif [ "$HTTP_CODE" = "302" ]; then
    # 本地没有，需要从 GitHub 下载
    echo "✗ Local package not found on server"
    echo "→ Downloading from GitHub (this may take a while or require network access)..."
    GITHUB_REPO="gocronx-team/gocron"
    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/gocron-node-${OS}-${ARCH}.zip"
    else
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/gocron-node-${OS}-${ARCH}.tar.gz"
    fi
else
    echo "✗ Failed to check server status (HTTP $HTTP_CODE)"
    echo "→ Trying GitHub as fallback..."
    GITHUB_REPO="gocronx-team/gocron"
    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/gocron-node-${OS}-${ARCH}.zip"
    else
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/gocron-node-${OS}-${ARCH}.tar.gz"
    fi
fi

echo "Downloading from: $DOWNLOAD_URL"
if [ "$OS" = "windows" ]; then
    curl -fsSL "$DOWNLOAD_URL" -o gocron-node.zip
    unzip -q gocron-node.zip
else
    curl -fsSL "$DOWNLOAD_URL" -o gocron-node.tar.gz
    tar -xzf gocron-node.tar.gz
fi

sudo mkdir -p "$INSTALL_DIR"
sudo cp -r gocron-node*/* "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/gocron-node"

echo "Registering agent..."
# 获取本机IP地址，如果失败则使用hostname
if [ "$OS" = "darwin" ]; then
    HOSTNAME=$(ipconfig getifaddr en0 2>/dev/null || hostname)
elif [ "$OS" = "linux" ]; then
    HOSTNAME=$(hostname -I 2>/dev/null | awk '{print $1}' || hostname)
else
    HOSTNAME=$(hostname)
fi
echo "Using hostname/IP: $HOSTNAME"
REGISTER_URL="${GOCRON_SERVER}/api/agent/register"
RESPONSE=$(curl -fsSL -X POST "$REGISTER_URL" \
    -H "Content-Type: application/json" \
    -d "{\"token\":\"$TOKEN\",\"hostname\":\"$HOSTNAME\"}")

if echo "$RESPONSE" | grep -q '"code":0'; then
    echo "Agent registered successfully"
else
    echo "Failed to register agent: $RESPONSE"
    exit 1
fi

if [ "$OS" = "linux" ]; then
    sudo tee /etc/systemd/system/${SERVICE_NAME}.service > /dev/null <<EOF
[Unit]
Description=Gocron Node Agent
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/gocron-node
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    sudo systemctl daemon-reload
    sudo systemctl enable ${SERVICE_NAME}
    sudo systemctl start ${SERVICE_NAME}
    echo "Service status:"
    sudo systemctl status ${SERVICE_NAME} --no-pager
elif [ "$OS" = "darwin" ]; then
    echo "macOS detected. Starting gocron-node..."
    # 先停止已存在的进程
    pkill -f gocron-node 2>/dev/null || true
    sleep 1
    nohup $INSTALL_DIR/gocron-node > /tmp/gocron-node.log 2>&1 &
    echo "gocron-node started in background (PID: $!)"
    echo "Log file: /tmp/gocron-node.log"
fi

cd /
rm -rf "$TMP_DIR"

echo ""
echo "========================================"
echo "Installation completed successfully!"
echo "========================================"
echo ""
echo "Agent Management Commands:"
echo ""
if [ "$OS" = "linux" ]; then
    echo "  Start:   sudo systemctl start ${SERVICE_NAME}"
    echo "  Stop:    sudo systemctl stop ${SERVICE_NAME}"
    echo "  Restart: sudo systemctl restart ${SERVICE_NAME}"
    echo "  Status:  sudo systemctl status ${SERVICE_NAME}"
    echo "  Logs:    sudo journalctl -u ${SERVICE_NAME} -f"
    echo ""
    echo "Uninstall:"
    echo "  sudo systemctl stop ${SERVICE_NAME}"
    echo "  sudo systemctl disable ${SERVICE_NAME}"
    echo "  sudo rm /etc/systemd/system/${SERVICE_NAME}.service"
    echo "  sudo systemctl daemon-reload"
    echo "  sudo rm -rf ${INSTALL_DIR}"
elif [ "$OS" = "darwin" ]; then
    echo "  Stop:    pkill -f gocron-node"
    echo "  Start:   nohup ${INSTALL_DIR}/gocron-node > /tmp/gocron-node.log 2>&1 &"
    echo "  Logs:    tail -f /tmp/gocron-node.log"
    echo "  Status:  ps aux | grep gocron-node | grep -v grep"
    echo ""
    echo "Uninstall:"
    echo "  pkill -f gocron-node"
    echo "  sudo rm -rf ${INSTALL_DIR}"
    echo "  rm /tmp/gocron-node.log"
fi
echo ""
echo "Installation directory: ${INSTALL_DIR}"
echo "========================================"
`

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(script))
}

// Register agent注册
func Register(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Hostname string `json:"hostname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondError(c, "Invalid request", err)
		return
	}

	agentToken := &models.AgentToken{}
	if err := agentToken.FindByToken(req.Token); err != nil {
		base.RespondError(c, "Invalid token")
		return
	}

	if time.Now().After(agentToken.ExpiresAt) {
		base.RespondError(c, "Token expired")
		return
	}
	if agentToken.Used {
		base.RespondError(c, "Token already used")
		return
	}

	// 原子地领取 token：用 WHERE used=false 的条件更新避免并发复用竞态
	claim := models.Db.Model(&models.AgentToken{}).
		Where("token = ? AND used = ?", req.Token, false).
		Updates(map[string]interface{}{"used": true, "used_at": time.Now()})
	if claim.Error != nil {
		base.RespondError(c, "Operation failed", claim.Error)
		return
	}
	if claim.RowsAffected == 0 {
		base.RespondError(c, "Token already used")
		return
	}

	host := &models.Host{
		Name:   req.Hostname,
		Alias:  req.Hostname,
		Port:   5921,
		Remark: "Auto registered",
	}

	exists, err := host.NameExists(req.Hostname, 0)
	if err != nil {
		logger.Error("检查主机是否存在失败:", err)
		base.RespondError(c, "Operation failed", err)
		return
	}

	if !exists {
		if _, err := host.Create(); err != nil {
			logger.Error("创建主机失败:", err)
			base.RespondError(c, "Failed to create host", err)
			return
		}
		logger.Infof("主机注册成功: %s", req.Hostname)
	} else {
		logger.Infof("主机已存在，跳过创建: %s", req.Hostname)
	}

	base.RespondSuccess(c, "Registration successful", nil)
}

// Download 优先从本地 gocron-node-package 目录下载，如果不存在则重定向到 GitHub Release
func Download(c *gin.Context) {
	osName := c.Query("os")
	arch := c.Query("arch")

	if osName == "" || arch == "" {
		c.String(http.StatusBadRequest, "os and arch are required")
		return
	}

	// 安全检查: 白名单验证，映射为内部安全常量字符串以消除路径注入风险
	var safeOS string
	switch osName {
	case "linux":
		safeOS = "linux"
	case "darwin":
		safeOS = "darwin"
	case "windows":
		safeOS = "windows"
	default:
		logger.Warnf("非法的 os 参数: %s", osName)
		c.String(http.StatusBadRequest, "invalid os parameter")
		return
	}

	var safeArch string
	switch arch {
	case "amd64":
		safeArch = "amd64"
	case "arm64":
		safeArch = "arm64"
	case "386":
		safeArch = "386"
	default:
		logger.Warnf("非法的 arch 参数: %s", arch)
		c.String(http.StatusBadRequest, "invalid arch parameter")
		return
	}

	// 根据操作系统选择文件扩展名
	ext := ".tar.gz"
	if safeOS == "windows" {
		ext = ".zip"
	}

	filename := fmt.Sprintf("gocron-node-%s-%s%s", safeOS, safeArch, ext)

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		logger.Errorf("获取可执行文件路径失败: %v", err)
		// 降级到 GitHub
		githubURL := fmt.Sprintf("https://github.com/gocronx-team/gocron/releases/latest/download/%s", filename)
		logger.Warnf("✗ 无法获取可执行文件路径，重定向到 GitHub: %s", githubURL)
		c.Redirect(http.StatusFound, githubURL)
		return
	}

	execDir := filepath.Dir(execPath)

	// 优先检查本地 gocron-node-package 目录（相对于可执行文件所在目录）
	packageDir := filepath.Join(execDir, "gocron-node-package")
	localPath := filepath.Join(packageDir, filename)

	logger.Infof("下载请求: os=%s, arch=%s, 查找路径: %s", safeOS, safeArch, localPath)

	// 安全检查: 确保最终路径在 packageDir 内,防止路径遍历
	cleanPath := filepath.Clean(localPath)
	cleanPackageDir := filepath.Clean(packageDir)

	// 使用 filepath.Rel 检查路径关系
	rel, err := filepath.Rel(cleanPackageDir, cleanPath)
	if err != nil || len(rel) > 0 && (rel[0] == '.' && len(rel) > 1 && rel[1] == '.') {
		logger.Warnf("检测到路径遍历攻击尝试: %s (相对路径: %s)", localPath, rel)
		c.String(http.StatusBadRequest, "invalid file path")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(cleanPath); err == nil {
		logger.Infof("✓ 本地安装包存在，提供文件: %s", cleanPath)
		c.File(cleanPath)
		return
	}

	// 本地文件不存在，重定向到 GitHub Release
	githubURL := fmt.Sprintf("https://github.com/gocronx-team/gocron/releases/latest/download/%s", filename)
	logger.Warnf("✗ 本地安装包不存在 (%s)，重定向到 GitHub: %s", localPath, githubURL)
	c.Redirect(http.StatusFound, githubURL)
}

func generateRandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getServerURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}
