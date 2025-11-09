package client

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
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
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Alias       string `json:"alias"`
	Version     string `json:"version"`
	NeedsCert   bool   `json:"needs_cert,omitempty"`
}

type RegisterResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type CertificateBundle struct {
	CACert     string `json:"ca_cert"`
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
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
		Hostname:  hostname,
		IP:        ip,
		Port:      port,
		Alias:     hostname,
		Version:   version,
		NeedsCert: !HasCertificates(), // 如果本地没有证书，请求服务端返回
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
	
	// 如果有证书，使用证书签名认证，否则使用 token
	var client *http.Client
	if HasCertificates() {
		// 读取证书并生成签名
		signature, err := generateCertSignature()
		if err != nil {
			return fmt.Errorf("generate cert signature failed: %v", err)
		}
		httpReq.Header.Set("X-Client-Cert-Signature", signature)
		client = &http.Client{Timeout: 10 * time.Second}
	} else {
		// 使用 token 认证
		httpReq.Header.Set("X-Register-Token", token)
		client = &http.Client{Timeout: 10 * time.Second}
	}
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

	// 如果服务端返回了证书，保存到本地
	if result.Data != nil {
		log.Debugf("Response data: %+v", result.Data)
		if certBundleData, ok := result.Data["cert_bundle"]; ok {
			log.Debug("Found cert_bundle in response")
			// 将 map 转换为 CertificateBundle
			certBundleJSON, _ := json.Marshal(certBundleData)
			var certBundle CertificateBundle
			if err := json.Unmarshal(certBundleJSON, &certBundle); err == nil {
				if err := saveCertificates(&certBundle); err != nil {
					log.Warnf("Failed to save certificates: %v", err)
				} else {
					log.Info("Client certificates saved successfully")
				}
			} else {
				log.Warnf("Failed to unmarshal cert bundle: %v", err)
			}
		} else {
			log.Debug("No cert_bundle in response data")
		}
	} else {
		log.Debug("Response data is nil")
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

// createMTLSClient 创建使用 mTLS 的 HTTP 客户端
func createMTLSClient() (*http.Client, error) {
	// 加载客户端证书和私钥
	cert, err := tls.LoadX509KeyPair("./certs/client.crt", "./certs/client.key")
	if err != nil {
		return nil, fmt.Errorf("load client certificate failed: %v", err)
	}

	// 加载 CA 证书
	caCert, err := os.ReadFile("./certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("load CA certificate failed: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// 配置 TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

// saveCertificates 保存证书到本地
func saveCertificates(bundle *CertificateBundle) error {
	certDir := "./certs"
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return fmt.Errorf("create cert directory failed: %v", err)
	}

	files := map[string]string{
		"ca.crt":     bundle.CACert,
		"client.crt": bundle.ClientCert,
		"client.key": bundle.ClientKey,
	}

	for filename, content := range files {
		path := fmt.Sprintf("%s/%s", certDir, filename)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return fmt.Errorf("write %s failed: %v", filename, err)
		}
	}

	return nil
}

// HasCertificates 检查本地是否已有证书
func HasCertificates() bool {
	requiredFiles := []string{"./certs/ca.crt", "./certs/client.crt", "./certs/client.key"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// generateCertSignature 生成证书签名（使用证书指纹）
func generateCertSignature() (string, error) {
	// 读取客户端证书
	certData, err := os.ReadFile("./certs/client.crt")
	if err != nil {
		return "", fmt.Errorf("read client cert failed: %v", err)
	}
	
	// 计算证书的 SHA256 指纹
	hash := sha256.Sum256(certData)
	signature := hex.EncodeToString(hash[:])
	
	return signature, nil
}
