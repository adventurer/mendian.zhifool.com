package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadReadsDatabasePasswordFromEnvironment(t *testing.T) {
	t.Setenv("MYSQL_TEST_PASSWORD", "test-password")
	t.Setenv("DB_PASSWORD", "")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte("" +
		"server:\n" +
		"  host: 127.0.0.1\n" +
		"  port: 8080\n" +
		"database:\n" +
		"  driver: mysql\n" +
		"  host: localhost\n" +
		"  port: 3306\n" +
		"  username: app\n" +
		"  password: file-password\n" +
		"  passwordEnv: MYSQL_TEST_PASSWORD\n" +
		"  name: mendian\n" +
		"  charset: utf8mb4\n" +
		"  parseTime: true\n" +
		"  location: Local\n")
	if err := os.WriteFile(configPath, contents, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Database.Password != "test-password" {
		t.Fatalf("expected password from environment, got %q", cfg.Database.Password)
	}
	if cfg.Database.Driver != "mysql" || cfg.Database.Name != "mendian" {
		t.Fatalf("unexpected database config: %+v", cfg.Database)
	}
}

func TestWeChatPayConfigRequiresAllCredentials(t *testing.T) {
	cfg := WeChatPayConfig{
		MerchantID:        "merchant",
		AppID:             "app",
		AppSecret:         "secret",
		MchID:             "mch",
		CertificateSerial: "serial",
		APIv3Key:          "api-v3-key",
		PrivateKeyPath:    "private-key.pem",
		NotifyURL:         "https://pay.example.com/api/payment/wechat/notify/merchant",
	}
	if !cfg.Configured() {
		t.Fatal("expected complete WeChat Pay config to be enabled")
	}
	cfg.APIv3Key = ""
	if cfg.Configured() {
		t.Fatal("expected incomplete WeChat Pay config to remain disabled")
	}
}

func TestWeChatPayConfigSupportsMultipleMerchants(t *testing.T) {
	contents := []byte("wechatPay:\n" +
		"  merchants:\n" +
		"    merchant-a:\n" +
		"      appId: app-a\n" +
		"      appSecret: secret-a\n" +
		"      mchId: mch-a\n" +
		"      certificateSerial: serial-a\n" +
		"      apiV3Key: 01234567890123456789012345678901\n" +
		"      privateKeyPath: a.pem\n" +
		"      notifyUrl: https://pay.example.com/api/payment/wechat/notify/merchant-a\n" +
		"    merchant-b:\n" +
		"      appId: app-b\n" +
		"      appSecret: secret-b\n" +
		"      mchId: mch-b\n" +
		"      certificateSerial: serial-b\n" +
		"      apiV3Key: 12345678901234567890123456789012\n" +
		"      privateKeyPath: b.pem\n" +
		"      notifyUrl: https://pay.example.com/api/payment/wechat/notify/merchant-b\n")
	var config struct {
		WeChatPay WeChatPayConfig `yaml:"wechatPay"`
	}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		t.Fatalf("unmarshal tenant-keyed WeChat Pay config: %v", err)
	}
	if got := config.WeChatPay.MerchantConfigs(); len(got) != 2 || got[0].MerchantID != "merchant-a" || got[1].MerchantID != "merchant-b" {
		t.Fatalf("unexpected merchant configs: %+v", got)
	}

	legacyConfig := WeChatPayConfig{MerchantID: "legacy", AppID: "app", AppSecret: "secret", MchID: "mch", CertificateSerial: "serial", APIv3Key: "api-v3-key", PrivateKeyPath: "private.pem", NotifyURL: "https://pay.example.com/api/payment/wechat/notify/legacy"}
	if got := legacyConfig.MerchantConfigs(); len(got) != 1 || got[0].MerchantID != "legacy" {
		t.Fatalf("expected legacy config to resolve as one merchant, got %+v", got)
	}
}

func TestConfigRejectsMismatchedWeChatPayMerchantIDKey(t *testing.T) {
	merchant := WeChatPayConfig{MerchantID: "different", AppID: "app", AppSecret: "secret", MchID: "mch", CertificateSerial: "serial", APIv3Key: "01234567890123456789012345678901", PrivateKeyPath: "private.pem", NotifyURL: "https://pay.example.com/api/payment/wechat/notify/different"}
	cfg := Config{
		Server:    ServerConfig{Host: "127.0.0.1", Port: 8080},
		Database:  DatabaseConfig{Driver: "mysql", Host: "localhost", Port: 3306, Username: "app", Name: "mendian", Charset: "utf8mb4"},
		WeChatPay: WeChatPayConfig{Merchants: map[string]WeChatPayConfig{"tenant-id": merchant}},
	}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected WeChat Pay merchant ID/key mismatch to be rejected")
	}
}

func TestServerTLSModesAreMutuallyExclusive(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Host:          "127.0.0.1",
			Port:          8081,
			TLSCert:       "server.crt",
			TLSKey:        "server.key",
			AutoTLSDomain: "api.example.com",
		},
		Database: DatabaseConfig{
			Driver: "mysql", Host: "localhost", Port: 3306,
			Username: "app", Name: "mendian", Charset: "utf8mb4",
		},
	}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected automatic and manual TLS configuration to conflict")
	}
}

func TestLoadReadsDatabasePasswordFromFile(t *testing.T) {
	t.Setenv("DB_PASSWORD", "")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte("" +
		"server:\n" +
		"  host: 127.0.0.1\n" +
		"  port: 8080\n" +
		"database:\n" +
		"  driver: mysql\n" +
		"  host: localhost\n" +
		"  port: 3306\n" +
		"  username: app\n" +
		"  password: file-password\n" +
		"  name: mendian\n" +
		"  charset: utf8mb4\n" +
		"  parseTime: true\n" +
		"  location: Local\n")
	if err := os.WriteFile(configPath, contents, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Database.Password != "file-password" {
		t.Fatal("expected password from config file")
	}
}
