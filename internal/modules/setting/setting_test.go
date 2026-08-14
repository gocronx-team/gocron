package setting

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestReadReturnsConfiguredValues(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.ini")
	content := `[default]
		db.engine=postgres
		db.host=10.0.0.1
		db.port=5432
		db.user=test_user
		db.password=test_pass
		db.database=test_db
		db.prefix=pre_
		db.charset=utf8mb4
		db.max.idle.conns=11
		db.max.open.conns=22
		allow_ips=127.0.0.1
		app.name=TestApp
		api.key=key
		api.secret=secret
		api.sign.enable=false
		concurrency.queue=200
		auth_secret=existing-secret
		enable_tls=false
    `
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	s, err := Read(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Db.Engine != "postgres" || s.Db.Host != "10.0.0.1" || s.Db.Port != 5432 {
		t.Fatalf("unexpected db config: %+v", s.Db)
	}
	if s.AppName != "TestApp" || s.ApiSignEnable {
		t.Fatalf("unexpected app config: %+v", s)
	}
	if s.ConcurrencyQueue != 200 || s.AuthSecret != "existing-secret" {
		t.Fatalf("unexpected concurrency/auth config: %+v", s)
	}
}

func TestReadGeneratesAuthSecretWhenMissing(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.ini")
	content := `[default]
db.engine=mysql
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	s, err := Read(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.AuthSecret == "" {
		t.Fatal("expected generated auth secret when config missing")
	}
}

func TestReadEnableTLSSucceedsWhenFilesExist(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	for _, p := range []string{caPath, certPath, keyPath} {
		if err := os.WriteFile(p, []byte("data"), 0o600); err != nil {
			t.Fatalf("failed to create tls file: %v", err)
		}
	}
	configPath := filepath.Join(dir, "app.ini")
	content := `[default]
enable_tls=true
ca_file=` + caPath + `
cert_file=` + certPath + `
key_file=` + keyPath + `
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	if _, err := Read(configPath); err != nil {
		t.Fatalf("expected tls config to be read successfully, got %v", err)
	}
}

func TestWriteValidatesArguments(t *testing.T) {
	if err := Write(nil, ""); err == nil {
		t.Fatal("expected error for empty config")
	}
	if err := Write([]string{"key"}, ""); err == nil {
		t.Fatal("expected error for odd number of config entries")
	}
}

func TestWritePersistsKeyValuePairs(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.ini")
	data := []string{
		"db.engine", "sqlite",
		"db.host", "",
		"api.sign.enable", "false",
	}

	if err := Write(data, configPath); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	section := cfg.Section(DefaultSection)
	if section.Key("db.engine").String() != "sqlite" {
		t.Fatalf("db.engine mismatch, got %s", section.Key("db.engine").String())
	}
	if section.Key("api.sign.enable").String() != "false" {
		t.Fatalf("api.sign.enable mismatch, got %s", section.Key("api.sign.enable").String())
	}
}

func TestWriteSetsSecureFilePermissions(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.ini")
	data := []string{
		"db.password", "secret123",
		"auth_secret", "token456",
	}

	if err := Write(data, configPath); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Fatalf("expected file permission 0600, got %#o", perm)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("GOCRON_DB_ENGINE", "postgres")
	t.Setenv("GOCRON_DB_HOST", "postgresql.default.svc")
	t.Setenv("GOCRON_DB_PORT", "5432")
	t.Setenv("GOCRON_DB_PASSWORD", "database-secret")
	t.Setenv("GOCRON_AUTH_SECRET", "shared-auth-secret")

	var config Setting
	if err := ApplyEnvOverrides(&config); err != nil {
		t.Fatalf("ApplyEnvOverrides() error = %v", err)
	}
	if config.Db.Engine != "postgres" || config.Db.Host != "postgresql.default.svc" || config.Db.Port != 5432 {
		t.Fatalf("unexpected database overrides: %+v", config.Db)
	}
	if config.Db.Password != "database-secret" || config.AuthSecret != "shared-auth-secret" {
		t.Fatal("managed secrets were not applied")
	}
}

func TestApplyEnvOverridesRejectsInvalidPort(t *testing.T) {
	t.Setenv("GOCRON_DB_PORT", "invalid")
	if err := ApplyEnvOverrides(&Setting{}); err == nil {
		t.Fatal("expected invalid port to be rejected")
	}
}

func TestClearEnvOverrides(t *testing.T) {
	t.Setenv("GOCRON_DB_PASSWORD", "database-secret")
	t.Setenv("GOCRON_ADMIN_PASSWORD", "admin-secret")
	ClearEnvOverrides()
	if os.Getenv("GOCRON_DB_PASSWORD") != "" || os.Getenv("GOCRON_ADMIN_PASSWORD") != "" {
		t.Fatal("managed credentials were not cleared")
	}
}
