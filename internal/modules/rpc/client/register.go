package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type RegisterRequest struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Alias    string `json:"alias"`
	Version  string `json:"version"`
}

type RegisterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RegisterToServer 向 gocron 服务端注册
func RegisterToServer(serverURL string, port int, version string, token string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get hostname failed: %v", err)
	}

	// 获取本机 IP
	ip, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("get local IP failed: %v", err)
	}

	req := RegisterRequest{
		Hostname: hostname,
		IP:       ip,
		Port:     port,
		Alias:    hostname,
		Version:  version,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %v", err)
	}

	url := fmt.Sprintf("%s/api/host/register", serverURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Register-Token", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %v", err)
	}

	var result RegisterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response failed: %v", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("register failed: %s", result.Message)
	}

	log.Infof("Successfully registered to %s (IP: %s, Port: %d)", serverURL, ip, port)
	return nil
}

// StartHeartbeat 启动心跳，定期向服务端注册
func StartHeartbeat(serverURL string, port int, version string, token string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 首次立即注册
	if err := RegisterToServer(serverURL, port, version, token); err != nil {
		log.Errorf("Initial registration failed: %v", err)
	}

	// 定期心跳
	for range ticker.C {
		if err := RegisterToServer(serverURL, port, version, token); err != nil {
			log.Errorf("Heartbeat registration failed: %v", err)
		}
	}
}

// getLocalIP 获取本机非回环 IP 地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IP address found")
}
