package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := deriveKey("test-master-key")
	cases := []string{"", "hello", "p@ssw0rd!#$", strings.Repeat("x", 4096), "多字节密钥🔐"}
	for _, plain := range cases {
		ciphertext, err := encrypt(key[:], plain)
		if err != nil {
			t.Fatalf("encrypt(%q) error: %v", plain, err)
		}
		if ciphertext == plain && plain != "" {
			t.Errorf("ciphertext equals plaintext for %q", plain)
		}
		got, err := decrypt(key[:], ciphertext)
		if err != nil {
			t.Fatalf("decrypt error: %v", err)
		}
		if got != plain {
			t.Errorf("round trip mismatch: got %q want %q", got, plain)
		}
	}
}

func TestEncryptUsesRandomNonce(t *testing.T) {
	key := deriveKey("k")
	a, _ := encrypt(key[:], "same")
	b, _ := encrypt(key[:], "same")
	if a == b {
		t.Error("two encryptions of same plaintext produced identical ciphertext (nonce not random)")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	k1 := deriveKey("key-one")
	k2 := deriveKey("key-two")
	ciphertext, _ := encrypt(k1[:], "secret")
	if _, err := decrypt(k2[:], ciphertext); err == nil {
		t.Error("expected error decrypting with wrong key, got nil")
	}
}

func TestDecryptTamperedCiphertextFails(t *testing.T) {
	key := deriveKey("k")
	ciphertext, _ := encrypt(key[:], "secret")
	raw, _ := base64.StdEncoding.DecodeString(ciphertext)
	raw[len(raw)-1] ^= 0xFF // 翻转 GCM tag 末字节
	tampered := base64.StdEncoding.EncodeToString(raw)
	if _, err := decrypt(key[:], tampered); err == nil {
		t.Error("expected error decrypting tampered ciphertext, got nil")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	key := deriveKey("k")
	if _, err := decrypt(key[:], "not-base64!!!"); err != ErrInvalidCiphertext {
		t.Errorf("expected ErrInvalidCiphertext for bad base64, got %v", err)
	}
	if _, err := decrypt(key[:], "AAAA"); err != ErrInvalidCiphertext {
		t.Errorf("expected ErrInvalidCiphertext for too-short input, got %v", err)
	}
}

func TestNotConfigured(t *testing.T) {
	saved := masterKey
	masterKey = nil
	defer func() { masterKey = saved }()

	if Configured() {
		t.Error("Configured() should be false when masterKey is nil")
	}
	if _, err := Encrypt("x"); err != ErrNotConfigured {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
	if _, err := Decrypt("x"); err != ErrNotConfigured {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

func TestPackageLevelEncryptDecrypt(t *testing.T) {
	saved := masterKey
	key := deriveKey("pkg-level")
	masterKey = key[:]
	defer func() { masterKey = saved }()

	ciphertext, err := Encrypt("hello")
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}
	got, err := Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q want hello", got)
	}
}
