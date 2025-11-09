// Command gocron-node
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/gocronx-team/gocron/internal/modules/rpc/auth"
	"github.com/gocronx-team/gocron/internal/modules/rpc/server"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	log "github.com/sirupsen/logrus"
)

var (
	AppVersion, BuildDate, GitCommit string
)

func main() {
	var serverAddr string
	var allowRoot bool
	var version bool
	var CAFile string
	var certFile string
	var keyFile string
	var enableTLS bool
	var logLevel string
	var gocronURL string
	var enableRegister bool
	var registerToken string
	flag.BoolVar(&allowRoot, "allow-root", false, "./gocron-node -allow-root")
	flag.StringVar(&serverAddr, "s", "0.0.0.0:5921", "./gocron-node -s ip:port")
	flag.BoolVar(&version, "v", false, "./gocron-node -v")
	flag.BoolVar(&enableTLS, "enable-tls", false, "./gocron-node -enable-tls")
	flag.StringVar(&CAFile, "ca-file", "", "./gocron-node -ca-file path")
	flag.StringVar(&certFile, "cert-file", "", "./gocron-node -cert-file path")
	flag.StringVar(&keyFile, "key-file", "", "./gocron-node -key-file path")
	flag.StringVar(&logLevel, "log-level", "info", "-log-level error")
	flag.StringVar(&gocronURL, "gocron-url", "", "./gocron-node -gocron-url http://gocron-server:5920")
	flag.BoolVar(&enableRegister, "enable-register", false, "./gocron-node -enable-register")
	flag.StringVar(&registerToken, "register-token", "", "./gocron-node -register-token <token>")
	flag.Parse()
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)

	if version {
		utils.PrintAppVersion(AppVersion, GitCommit, BuildDate)
		return
	}

	if enableTLS {
		if !utils.FileExist(CAFile) {
			log.Fatalf("failed to read ca cert file: %s", CAFile)
		}
		if !utils.FileExist(certFile) {
			log.Fatalf("failed to read server cert file: %s", certFile)
			return
		}
		if !utils.FileExist(keyFile) {
			log.Fatalf("failed to read server key file: %s", keyFile)
			return
		}
	}

	certificate := auth.Certificate{
		CAFile:   strings.TrimSpace(CAFile),
		CertFile: strings.TrimSpace(certFile),
		KeyFile:  strings.TrimSpace(keyFile),
	}

	if runtime.GOOS != "windows" && os.Getuid() == 0 && !allowRoot {
		log.Fatal("Do not run gocron-node as root user")
		return
	}

	// 启动自动注册
	if enableRegister && gocronURL != "" {
		// 解析端口
		var port int
		if _, err := fmt.Sscanf(serverAddr, "%*[^:]:%d", &port); err != nil {
			port = 5921
		}
		
		// 检查是否已有证书，如果有则使用证书认证，否则使用 token
		if server.HasCertificates() {
			log.Info("Using mTLS authentication")
			go server.StartAutoRegister(gocronURL, port, AppVersion, "")
		} else {
			if registerToken == "" {
				log.Fatal("register-token is required for initial registration")
			}
			log.Info("Using token authentication for initial registration")
			go server.StartAutoRegister(gocronURL, port, AppVersion, registerToken)
		}
	}

	server.Start(serverAddr, enableTLS, certificate)
}
