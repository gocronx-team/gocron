package ca

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"sync"
	"testing"
	
	"github.com/gocronx-team/gocron/internal/modules/logger"
)

func init() {
	// 初始化 logger 用于测试
	logger.InitLogger()
}

func TestLoadOrCreateCA(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	
	// 第一次调用应该创建新的 CA
	ca1, err := LoadOrCreateCA()
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}
	
	if ca1.CACert == nil || ca1.CAKey == nil {
		t.Fatal("CA certificate or key is nil")
	}
	
	// 验证文件已创建
	caDir := filepath.Join(tmpDir, ".gocron", "ca")
	caCertPath := filepath.Join(caDir, "ca.crt")
	caKeyPath := filepath.Join(caDir, "ca.key")
	clientCertPath := filepath.Join(caDir, "client.crt")
	clientKeyPath := filepath.Join(caDir, "client.key")
	
	if !fileExists(caCertPath) {
		t.Fatal("CA certificate file not created")
	}
	if !fileExists(caKeyPath) {
		t.Fatal("CA key file not created")
	}
	if !fileExists(clientCertPath) {
		t.Fatal("Client certificate file not created")
	}
	if !fileExists(clientKeyPath) {
		t.Fatal("Client key file not created")
	}
	
	// 第二次调用应该加载现有的 CA
	ca2, err := LoadOrCreateCA()
	if err != nil {
		t.Fatalf("Failed to load CA: %v", err)
	}
	
	// 验证加载的 CA 与创建的 CA 相同
	if ca1.CACert.SerialNumber.Cmp(ca2.CACert.SerialNumber) != 0 {
		t.Fatal("Loaded CA is different from created CA")
	}
}

func TestGenerateServerCert(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	
	ca, err := LoadOrCreateCA()
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}
	
	// 生成服务端证书
	certPEM, keyPEM, err := ca.GenerateServerCert("192.168.1.100", "test-node")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}
	
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		t.Fatal("Generated certificate or key is empty")
	}
	
	// 验证证书可以被加载
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("Failed to load generated certificate: %v", err)
	}
	
	// 验证证书是由 CA 签发的
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}
	
	// 验证证书的 CN
	if x509Cert.Subject.CommonName != "192.168.1.100" {
		t.Fatalf("Certificate CN mismatch: got %s, want 192.168.1.100", x509Cert.Subject.CommonName)
	}
	
	// 验证证书用途
	hasServerAuth := false
	hasClientAuth := false
	for _, usage := range x509Cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			hasServerAuth = true
		}
		if usage == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
		}
	}
	if !hasServerAuth {
		t.Fatal("Certificate missing ServerAuth usage")
	}
	if !hasClientAuth {
		t.Fatal("Certificate missing ClientAuth usage")
	}
}

func TestGenerateClientCert(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	
	ca, err := LoadOrCreateCA()
	if err != nil {
		t.Fatalf("Failed to create CA: %v", err)
	}
	
	// 生成客户端证书
	certPEM, keyPEM, err := ca.GenerateClientCert()
	if err != nil {
		t.Fatalf("Failed to generate client certificate: %v", err)
	}
	
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		t.Fatal("Generated certificate or key is empty")
	}
	
	// 验证证书可以被加载
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("Failed to load generated certificate: %v", err)
	}
	
	// 验证证书是由 CA 签发的
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}
	
	// 验证证书的 CN
	if x509Cert.Subject.CommonName != "Gocron Client" {
		t.Fatalf("Certificate CN mismatch: got %s, want Gocron Client", x509Cert.Subject.CommonName)
	}
	
	// 验证证书用途
	hasClientAuth := false
	for _, usage := range x509Cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
		}
	}
	if !hasClientAuth {
		t.Fatal("Certificate missing ClientAuth usage")
	}
}

func TestGetGlobalCA(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	
	// 重置全局 CA
	globalCA = nil
	globalCAOnce = *new(sync.Once)
	
	// 第一次调用
	ca1 := GetGlobalCA()
	if ca1 == nil {
		t.Fatal("GetGlobalCA returned nil")
	}
	
	// 第二次调用应该返回相同的实例
	ca2 := GetGlobalCA()
	if ca1 != ca2 {
		t.Fatal("GetGlobalCA returned different instances")
	}
}
