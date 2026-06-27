// Package crypto 提供机密(Secret)的对称加密能力。
//
// 设计要点:
//   - 主密钥来自环境变量 GOCRON_SECRET_KEY,经 SHA-256 派生为 32 字节,
//     用于 AES-256-GCM。读取后立即从环境清除,降低被其它进程/子任务读到的风险。
//   - GCM 自带认证,密文被篡改会在解密时报错。
//   - 每次加密使用随机 nonce,前置到密文一起 base64 编码存储。
//   - 未配置主密钥时 Encrypt/Decrypt 返回 ErrNotConfigured,由上层给出友好提示。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// SecretKeyEnv 是配置主密钥的环境变量名。
const SecretKeyEnv = "GOCRON_SECRET_KEY"

// ErrNotConfigured 表示尚未配置主密钥,无法加解密。
var ErrNotConfigured = errors.New("secret encryption key not configured (set GOCRON_SECRET_KEY)")

// ErrInvalidCiphertext 表示密文格式非法或长度不足。
var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// masterKey 是派生后的 32 字节 AES 密钥;nil 表示未配置。
var masterKey []byte

// Init 从环境变量读取主密钥并派生 AES 密钥,随后清除该环境变量。
// 未设置环境变量时不报错——机密功能保持不可用,直到配置为止。
func Init() {
	raw := os.Getenv(SecretKeyEnv)
	if raw == "" {
		return
	}
	key := deriveKey(raw)
	masterKey = key[:]
	_ = os.Unsetenv(SecretKeyEnv)
}

// Configured 返回主密钥是否已就绪。
func Configured() bool {
	return masterKey != nil
}

// Encrypt 用主密钥加密明文,返回 base64(nonce||ciphertext)。
func Encrypt(plaintext string) (string, error) {
	if masterKey == nil {
		return "", ErrNotConfigured
	}
	return encrypt(masterKey, plaintext)
}

// Decrypt 解密 Encrypt 产生的密文。
func Decrypt(ciphertext string) (string, error) {
	if masterKey == nil {
		return "", ErrNotConfigured
	}
	return decrypt(masterKey, ciphertext)
}

// deriveKey 把任意长度的口令派生为固定 32 字节密钥。
func deriveKey(secret string) [32]byte {
	return sha256.Sum256([]byte(secret))
}

// encrypt 用给定密钥执行 AES-256-GCM 加密。独立于全局状态,便于测试。
func encrypt(key []byte, plaintext string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// decrypt 是 encrypt 的逆操作。
func decrypt(key []byte, ciphertext string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", ErrInvalidCiphertext
	}
	nonce, sealed := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
