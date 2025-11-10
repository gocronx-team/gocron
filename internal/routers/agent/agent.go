package agent

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
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
		json := utils.JsonResponse{}
		c.String(http.StatusOK, json.CommonFailure(i18n.T(c, "operation_failed"), err))
		return
	}

	serverURL := getServerURL(c)
	installCmdLinux := fmt.Sprintf("curl -fsSL %s/api/agent/install.sh | bash -s -- %s", serverURL, token)
	installCmdWindows := fmt.Sprintf("iwr -useb %s/api/agent/install.ps1?token=%s | iex", serverURL, token)

	json := utils.JsonResponse{}
	c.String(http.StatusOK, json.Success(i18n.T(c, "operation_success"), map[string]interface{}{
		"token":              token,
		"expires_at":         expiresAt,
		"install_cmd":        installCmdLinux,
		"install_cmd_windows": installCmdWindows,
	}))
}

// InstallScript 返回安装脚本
func InstallScript(c *gin.Context) {
	script := `#!/bin/bash
set -e

TOKEN="$1"
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

DOWNLOAD_URL="${GOCRON_SERVER}/api/agent/download?os=${OS}&arch=${ARCH}"
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo "Downloading from $DOWNLOAD_URL..."
curl -fsSL "$DOWNLOAD_URL" -o gocron-node.tar.gz
tar -xzf gocron-node.tar.gz

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
User=root
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

// InstallScriptWindows 返回Windows PowerShell安装脚本
func InstallScriptWindows(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Token is required")
		return
	}

	script := `$ErrorActionPreference = "Stop"

$TOKEN = "` + token + `"
$GOCRON_SERVER = "` + getServerURL(c) + `"
$INSTALL_DIR = "C:\Program Files\gocron-node"
$SERVICE_NAME = "gocron-node"

Write-Host "Installing gocron-node for Windows..."

$ARCH = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$DOWNLOAD_URL = "${GOCRON_SERVER}/api/agent/download?os=windows&arch=${ARCH}"

$TMP_DIR = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $TMP_DIR | Out-Null
Set-Location $TMP_DIR

Write-Host "Downloading from $DOWNLOAD_URL..."
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "gocron-node.zip"
Expand-Archive -Path "gocron-node.zip" -DestinationPath .

if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
}
Copy-Item -Path "gocron-node*\*" -Destination $INSTALL_DIR -Recurse -Force

Write-Host "Registering agent..."
$HOSTNAME = $env:COMPUTERNAME
Write-Host "Using hostname: $HOSTNAME"

$REGISTER_URL = "${GOCRON_SERVER}/api/agent/register"
$body = @{
    token = $TOKEN
    hostname = $HOSTNAME
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri $REGISTER_URL -Method Post -Body $body -ContentType "application/json"
if ($response.code -eq 0) {
    Write-Host "Agent registered successfully"
} else {
    Write-Host "Failed to register agent: $($response.message)"
    exit 1
}

Write-Host "Creating Windows service..."
$servicePath = "$INSTALL_DIR\gocron-node.exe"
if (Get-Service -Name $SERVICE_NAME -ErrorAction SilentlyContinue) {
    Stop-Service -Name $SERVICE_NAME -Force
    sc.exe delete $SERVICE_NAME
    Start-Sleep -Seconds 2
}

sc.exe create $SERVICE_NAME binPath= $servicePath start= auto
sc.exe description $SERVICE_NAME "Gocron Node Agent"
Start-Service -Name $SERVICE_NAME

Set-Location $env:TEMP
Remove-Item -Path $TMP_DIR -Recurse -Force

Write-Host ""
Write-Host "========================================"
Write-Host "Installation completed successfully!"
Write-Host "========================================"
Write-Host ""
Write-Host "Agent Management Commands:"
Write-Host "  Start:   Start-Service -Name $SERVICE_NAME"
Write-Host "  Stop:    Stop-Service -Name $SERVICE_NAME"
Write-Host "  Restart: Restart-Service -Name $SERVICE_NAME"
Write-Host "  Status:  Get-Service -Name $SERVICE_NAME"
Write-Host ""
Write-Host "Uninstall:"
Write-Host "  Stop-Service -Name $SERVICE_NAME -Force"
Write-Host "  sc.exe delete $SERVICE_NAME"
Write-Host "  Remove-Item -Path '$INSTALL_DIR' -Recurse -Force"
Write-Host ""
Write-Host "Installation directory: $INSTALL_DIR"
Write-Host "========================================"
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
		json := utils.JsonResponse{}
		c.String(http.StatusOK, json.CommonFailure("Invalid request", err))
		return
	}

	agentToken := &models.AgentToken{}
	if err := agentToken.FindByToken(req.Token); err != nil {
		json := utils.JsonResponse{}
		c.String(http.StatusOK, json.CommonFailure("Invalid token"))
		return
	}

	// 只检查是否过期，不检查是否已使用
	if time.Now().After(agentToken.ExpiresAt) {
		json := utils.JsonResponse{}
		c.String(http.StatusOK, json.CommonFailure("Token expired"))
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
		json := utils.JsonResponse{}
		c.String(http.StatusOK, json.CommonFailure("Operation failed", err))
		return
	}

	if !exists {
		if _, err := host.Create(); err != nil {
			logger.Error("创建主机失败:", err)
			json := utils.JsonResponse{}
			c.String(http.StatusOK, json.CommonFailure("Failed to create host", err))
			return
		}
		logger.Infof("主机注册成功: %s", req.Hostname)
	} else {
		logger.Infof("主机已存在，跳过创建: %s", req.Hostname)
	}

	json := utils.JsonResponse{}
	c.String(http.StatusOK, json.Success("Registration successful", nil))
}

// Download 下载agent二进制文件
func Download(c *gin.Context) {
	os := c.Query("os")
	arch := c.Query("arch")

	if os == "" || arch == "" {
		c.String(http.StatusBadRequest, "os and arch are required")
		return
	}

	// 查找匹配的包文件
	packagePattern := fmt.Sprintf("./gocron-node-package/gocron-node-*-%s-%s.tar.gz", os, arch)
	matches, err := filepath.Glob(packagePattern)
	if err != nil || len(matches) == 0 {
		c.String(http.StatusNotFound, "Package not found")
		return
	}

	c.FileAttachment(matches[0], fmt.Sprintf("gocron-node-%s-%s.tar.gz", os, arch))
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getServerURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}
