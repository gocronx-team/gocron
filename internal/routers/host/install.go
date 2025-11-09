package host

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
)

// GetInstallScript 生成一键安装脚本
func GetInstallScript(c *gin.Context) {
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

	// 获取服务器地址
	serverURL := fmt.Sprintf("%s://%s", getScheme(c), c.Request.Host)
	
	// 检测操作系统（默认使用当前系统）
	os := c.DefaultQuery("os", "darwin")
	arch := c.DefaultQuery("arch", "arm64")
	
	script := generateInstallScript(serverURL, token, os, arch)
	
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, script)
}

func getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if scheme := c.GetHeader("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

func generateInstallScript(serverURL, token, os, arch string) string {
	if os == "windows" {
		return generateWindowsScript(serverURL, token, arch)
	}
	return generateUnixScript(serverURL, token, os, arch)
}

func generateUnixScript(serverURL, token, os, arch string) string {
	version := "v1.3.14" // 从实际版本号获取
	
	return fmt.Sprintf(`#!/bin/bash
set -e

echo "=========================================="
echo "  Gocron Node Auto-Install Script"
echo "=========================================="
echo ""

# 检测操作系统和架构
if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
else
    OS="linux"
fi

if [[ $(uname -m) == "arm64" ]] || [[ $(uname -m) == "aarch64" ]]; then
    ARCH="arm64"
else
    ARCH="amd64"
fi
VERSION="%s"
SERVER_URL="%s"
REGISTER_TOKEN="%s"

echo "OS: $OS"
echo "Architecture: $ARCH"
echo "Server: $SERVER_URL"
echo ""

# 创建安装目录
if [[ "$OS" == "darwin" ]]; then
    INSTALL_DIR="$HOME/gocron-node"
else
    INSTALL_DIR="/opt/gocron-node"
fi
echo "Creating installation directory: $INSTALL_DIR"
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 下载 gocron-node
echo "Downloading gocron-node..."
DOWNLOAD_URL="$SERVER_URL/download/agent?token=$REGISTER_TOKEN&os=$OS&arch=$ARCH"
if command -v curl &> /dev/null; then
    curl -L -o gocron-node.tar.gz "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -O gocron-node.tar.gz "$DOWNLOAD_URL"
else
    echo "Error: Neither wget nor curl is available"
    exit 1
fi

# 解压
echo "Extracting files..."
tar -xzf gocron-node.tar.gz
rm gocron-node.tar.gz

# 查找 gocron-node 可执行文件
BINARY=$(find . -name "gocron-node" -type f | head -n 1)
if [ -z "$BINARY" ]; then
    echo "Error: gocron-node binary not found"
    exit 1
fi

# 移动到安装目录
mv $BINARY ./gocron-node
chmod +x ./gocron-node

# 获取证书
echo "Fetching certificates..."
mkdir -p certs

# 获取本机 IP
if [[ "$OS" == "darwin" ]]; then
    LOCAL_IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -n 1)
else
    LOCAL_IP=$(hostname -I | awk '{print $1}')
fi

HOSTNAME=$(hostname)

if command -v curl &> /dev/null; then
    curl -s -X POST "$SERVER_URL/api/host/provision" \
      -H "Content-Type: application/json" \
      -H "X-Register-Token: $REGISTER_TOKEN" \
      -d '{"hostname":"'$HOSTNAME'","ip":"'$LOCAL_IP'"}' \
      -o certs/bundle.json
elif command -v wget &> /dev/null; then
    wget -q -O certs/bundle.json --header="Content-Type: application/json" \
      --header="X-Register-Token: $REGISTER_TOKEN" \
      --post-data='{"hostname":"'$HOSTNAME'","ip":"'$LOCAL_IP'"}' \
      "$SERVER_URL/api/host/provision"
fi

# 解析并保存证书
if command -v python3 &> /dev/null; then
    python3 -c "
import json, sys
with open('certs/bundle.json') as f:
    data = json.load(f)
    if data.get('code') == 0 and 'cert_bundle' in data.get('data', {}):
        bundle = data['data']['cert_bundle']
        with open('certs/ca.crt', 'w') as cf: cf.write(bundle['ca_cert'])
        with open('certs/server.crt', 'w') as cf: cf.write(bundle['server_cert'])
        with open('certs/server.key', 'w') as cf: cf.write(bundle['server_key'])
        print('Certificates saved successfully')
    else:
        msg = data.get('message', 'Unknown error')
        print('Failed to get certificates: ' + msg)
        sys.exit(1)
"
    if [ $? -ne 0 ]; then
        rm -f certs/bundle.json
        exit 1
    fi
fi
rm -f certs/bundle.json

# 根据操作系统创建服务
if [[ "$OS" == "darwin" ]]; then
    # macOS - 创建 launchd 服务
    echo "Creating launchd service..."
    PLIST_FILE="$HOME/Library/LaunchAgents/com.gocron.node.plist"
    mkdir -p "$HOME/Library/LaunchAgents"
    cat > "$PLIST_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.gocron.node</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/gocron-node</string>
        <string>-s</string>
        <string>0.0.0.0:5921</string>
        <string>-enable-tls</string>
        <string>-ca-file</string>
        <string>$INSTALL_DIR/certs/ca.crt</string>
        <string>-cert-file</string>
        <string>$INSTALL_DIR/certs/server.crt</string>
        <string>-key-file</string>
        <string>$INSTALL_DIR/certs/server.key</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$INSTALL_DIR</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$INSTALL_DIR/gocron-node.log</string>
    <key>StandardErrorPath</key>
    <string>$INSTALL_DIR/gocron-node.error.log</string>
</dict>
</plist>
EOF

    # 启动服务
    echo "Starting gocron-node service..."
    launchctl unload "$PLIST_FILE" 2>/dev/null || true
    launchctl load "$PLIST_FILE"
    
    echo ""
    echo "=========================================="
    echo "  Installation completed successfully!"
    echo "=========================================="
    echo ""
    echo "Useful commands:"
    echo "  - Check logs: tail -f $INSTALL_DIR/gocron-node.log"
    echo "  - Check errors: tail -f $INSTALL_DIR/gocron-node.error.log"
    echo "  - Restart: launchctl unload \"$PLIST_FILE\" && launchctl load \"$PLIST_FILE\""
    echo "  - Stop: launchctl unload \"$PLIST_FILE\""
else
    # Linux - 创建 systemd 服务
    echo "Creating systemd service..."
    sudo tee /etc/systemd/system/gocron-node.service > /dev/null <<EOF
[Unit]
Description=Gocron Node Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/gocron-node -s 0.0.0.0:5921 -enable-tls -ca-file $INSTALL_DIR/certs/ca.crt -cert-file $INSTALL_DIR/certs/server.crt -key-file $INSTALL_DIR/certs/server.key
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    # 启动服务
    echo "Starting gocron-node service..."
    sudo systemctl daemon-reload
    sudo systemctl enable gocron-node
    sudo systemctl start gocron-node
    
    echo ""
    echo "=========================================="
    echo "  Installation completed successfully!"
    echo "=========================================="
    echo ""
    echo "Service status:"
    sudo systemctl status gocron-node --no-pager
    echo ""
    echo "Useful commands:"
    echo "  - Check status: sudo systemctl status gocron-node"
    echo "  - View logs: sudo journalctl -u gocron-node -f"
    echo "  - Restart: sudo systemctl restart gocron-node"
    echo "  - Stop: sudo systemctl stop gocron-node"
fi
`, version, serverURL, token)
}

func generateWindowsScript(serverURL, token, arch string) string {
	version := "v1.3.14"
	
	return fmt.Sprintf(`@echo off
echo ==========================================
echo   Gocron Node Auto-Install Script
echo ==========================================
echo.

set ARCH=%s
set VERSION=%s
set SERVER_URL=%s
set REGISTER_TOKEN=%s

echo Architecture: %%ARCH%%
echo Server: %%SERVER_URL%%
echo.

REM Create installation directory
set INSTALL_DIR=C:\gocron-node
echo Creating installation directory: %%INSTALL_DIR%%
if not exist "%%INSTALL_DIR%%" mkdir "%%INSTALL_DIR%%"
cd /d "%%INSTALL_DIR%%"

REM Download gocron-node
echo Downloading gocron-node...
set DOWNLOAD_URL=%%SERVER_URL%%/download/agent?token=%%REGISTER_TOKEN%%^&os=windows^&arch=%%ARCH%%
powershell -Command "Invoke-WebRequest -Uri '%%DOWNLOAD_URL%%' -OutFile 'gocron-node.zip'"

REM Extract
echo Extracting files...
powershell -Command "Expand-Archive -Path 'gocron-node.zip' -DestinationPath '.' -Force"
del gocron-node.zip

REM Create start script
echo Creating start script...
(
echo @echo off
echo cd /d "%%INSTALL_DIR%%"
echo gocron-node.exe -s 0.0.0.0:5921 -enable-register -gocron-url %%SERVER_URL%% -register-token %%REGISTER_TOKEN%%
) > start-gocron-node.bat

echo.
echo ==========================================
echo   Installation completed!
echo ==========================================
echo.
echo To start gocron-node, run:
echo   %%INSTALL_DIR%%\start-gocron-node.bat
echo.
echo To install as Windows service, use NSSM:
echo   nssm install gocron-node "%%INSTALL_DIR%%\start-gocron-node.bat"
`, arch, version, serverURL, token)
}
