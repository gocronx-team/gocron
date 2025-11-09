package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gocronx-team/gocron/internal/modules/logger"
)

var (
	globalCA     *CertificateAuthority
	globalCAOnce sync.Once
)

type CertificateAuthority struct {
	CACert    *x509.Certificate
	CAKey     *rsa.PrivateKey
	CACertPEM []byte
	CAKeyPEM  []byte
}

// GetGlobalCA 获取全局 CA 实例
func GetGlobalCA() *CertificateAuthority {
	globalCAOnce.Do(func() {
		var err error
		globalCA, err = LoadOrCreateCA()
		if err != nil {
			logger.Fatalf("Failed to initialize CA: %v", err)
		}
	})
	return globalCA
}

// LoadOrCreateCA 加载或创建 CA 证书
func LoadOrCreateCA() (*CertificateAuthority, error) {
	caDir := filepath.Join(os.Getenv("HOME"), ".gocron", "ca")
	caCertPath := filepath.Join(caDir, "ca.crt")
	caKeyPath := filepath.Join(caDir, "ca.key")

	// 尝试加载现有 CA
	if fileExists(caCertPath) && fileExists(caKeyPath) {
		ca, err := loadCA(caCertPath, caKeyPath)
		if err == nil {
			logger.Info("Loaded existing CA certificate")
			// 确保客户端证书存在
			if err := ca.ensureClientCert(); err != nil {
				logger.Warnf("Failed to ensure client certificate: %v", err)
			}
			return ca, nil
		}
		logger.Warnf("Failed to load existing CA, creating new one: %v", err)
	}

	// 创建新 CA
	ca, err := createCA()
	if err != nil {
		return nil, err
	}

	// 保存 CA
	if err := os.MkdirAll(caDir, 0700); err != nil {
		return nil, err
	}

	if err := os.WriteFile(caCertPath, ca.CACertPEM, 0600); err != nil {
		return nil, err
	}

	if err := os.WriteFile(caKeyPath, ca.CAKeyPEM, 0600); err != nil {
		return nil, err
	}

	logger.Infof("Created new CA certificate at %s", caDir)
	
	// 生成客户端证书
	if err := ca.ensureClientCert(); err != nil {
		logger.Warnf("Failed to generate client certificate: %v", err)
	}
	
	return ca, nil
}

// ensureClientCert 确保客户端证书存在
func (ca *CertificateAuthority) ensureClientCert() error {
	caDir := filepath.Join(os.Getenv("HOME"), ".gocron", "ca")
	clientCertPath := filepath.Join(caDir, "client.crt")
	clientKeyPath := filepath.Join(caDir, "client.key")
	
	// 如果客户端证书已存在，跳过
	if fileExists(clientCertPath) && fileExists(clientKeyPath) {
		return nil
	}
	
	// 生成客户端证书
	clientCertPEM, clientKeyPEM, err := ca.GenerateClientCert()
	if err != nil {
		return err
	}
	
	// 保存客户端证书
	if err := os.WriteFile(clientCertPath, clientCertPEM, 0600); err != nil {
		return err
	}
	
	if err := os.WriteFile(clientKeyPath, clientKeyPEM, 0600); err != nil {
		return err
	}
	
	logger.Infof("Generated client certificate at %s", caDir)
	return nil
}

// createCA 创建新的 CA 证书
func createCA() (*CertificateAuthority, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Gocron CA",
			Organization: []string{"Gocron"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, err
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})

	return &CertificateAuthority{
		CACert:    caCert,
		CAKey:     caKey,
		CACertPEM: caCertPEM,
		CAKeyPEM:  caKeyPEM,
	}, nil
}

// loadCA 加载现有 CA 证书
func loadCA(certPath, keyPath string) (*CertificateAuthority, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, err
	}

	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, err
	}

	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return &CertificateAuthority{
		CACert:    caCert,
		CAKey:     caKey,
		CACertPEM: certPEM,
		CAKeyPEM:  keyPEM,
	}, nil
}

// GenerateServerCert 生成服务端证书（用于 agent 节点）
func (ca *CertificateAuthority) GenerateServerCert(ip, hostname string) (certPEM, keyPEM []byte, err error) {
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   ip,
			Organization: []string{"Gocron Node"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname, "localhost"},
		IPAddresses:           parseIPAddresses(ip),
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, &template, ca.CACert, &serverKey.PublicKey, ca.CAKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return certPEM, keyPEM, nil
}

// GenerateClientCert 生成客户端证书（用于 gocron 服务端连接 agent）
func (ca *CertificateAuthority) GenerateClientCert() (certPEM, keyPEM []byte, err error) {
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   "Gocron Client",
			Organization: []string{"Gocron"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, &template, ca.CACert, &clientKey.PublicKey, ca.CAKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	return certPEM, keyPEM, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// parseIPAddresses 解析 IP 地址，支持 127.0.0.1 和实际 IP
func parseIPAddresses(ip string) []net.IP {
	var ips []net.IP
	
	// 添加提供的 IP
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		ips = append(ips, parsedIP)
	}
	
	// 总是添加 localhost IPs
	ips = append(ips, net.ParseIP("127.0.0.1"))
	ips = append(ips, net.ParseIP("::1"))
	
	return ips
}

// GetClientCertPath 获取客户端证书路径
func GetClientCertPath() (certPath, keyPath string) {
	caDir := filepath.Join(os.Getenv("HOME"), ".gocron", "ca")
	return filepath.Join(caDir, "client.crt"), filepath.Join(caDir, "client.key")
}

// GetCACertPath 获取 CA 证书路径
func GetCACertPath() string {
	caDir := filepath.Join(os.Getenv("HOME"), ".gocron", "ca")
	return filepath.Join(caDir, "ca.crt")
}
