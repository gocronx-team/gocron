// Command gocron-node
package main

import (
	"flag"
	"os"
	"runtime"
	"strings"

	"github.com/gocronx-team/gocron/internal/modules/rpc/auth"
	"github.com/gocronx-team/gocron/internal/modules/rpc/server"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/memlimit"
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
	var token string
	flag.BoolVar(&allowRoot, "allow-root", false, "./gocron-node -allow-root")
	flag.StringVar(&serverAddr, "s", "0.0.0.0:5921", "./gocron-node -s ip:port")
	flag.BoolVar(&version, "v", false, "./gocron-node -v")
	flag.BoolVar(&enableTLS, "enable-tls", false, "./gocron-node -enable-tls")
	flag.StringVar(&CAFile, "ca-file", "", "./gocron-node -ca-file path")
	flag.StringVar(&certFile, "cert-file", "", "./gocron-node -cert-file path")
	flag.StringVar(&keyFile, "key-file", "", "./gocron-node -key-file path")
	flag.StringVar(&token, "token", "", "shared RPC token; when set, callers must present the same token")
	flag.StringVar(&logLevel, "log-level", "info", "-log-level error")
	flag.Parse()

	// 令牌优先用命令行,其次读环境变量,避免在进程列表中暴露。
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GOCRON_NODE_TOKEN"))
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)

	if version {
		utils.PrintAppVersion(AppVersion, GitCommit, BuildDate)
		return
	}

	// 容器内存感知:把 GOMEMLIMIT 设为 cgroup 内存上限的 90%,减少节点进程被 OOM kill。
	// 非容器/无限额环境为安全 no-op。
	if n, err := memlimit.SetFromCgroup(); err != nil {
		log.Warnf("memlimit: %v", err)
	} else if n > 0 {
		log.Infof("GOMEMLIMIT set from cgroup memory limit: %d bytes", n)
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

	server.Start(serverAddr, enableTLS, certificate, token)
}
