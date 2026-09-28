package payment

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mendian-backend/internal/config"
)

func TestNewGatewayRequires32ByteAPIv3Key(t *testing.T) {
	_, err := NewGateway(context.Background(), testPayConfig("short-key", "https://pay.example.com/api/payment/wechat/notify/demo"))
	if err == nil || !strings.Contains(err.Error(), "32 bytes") {
		t.Fatalf("expected API v3 key length error, got %v", err)
	}
}

func TestNewGatewayRequiresMatchingPublicNotifyPath(t *testing.T) {
	_, err := NewGateway(context.Background(), testPayConfig(strings.Repeat("k", 32), "https://pay.example.com/notify"))
	if err == nil || !strings.Contains(err.Error(), "path must match") {
		t.Fatalf("expected callback path error before loading private key, got %v", err)
	}
}

func TestNewGatewaySupportsWeChatPayPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal test private key: %v", err)
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal test public key: %v", err)
	}
	privateKeyPath := filepath.Join(t.TempDir(), "merchant-private.pem")
	publicKeyPath := filepath.Join(t.TempDir(), "wechat-public.pem")
	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}), 0600); err != nil {
		t.Fatalf("write test private key: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes}), 0600); err != nil {
		t.Fatalf("write test public key: %v", err)
	}

	cfg := testPayConfig(strings.Repeat("k", 32), "https://pay.example.com/api/payment/wechat/notify/demo")
	cfg.PrivateKeyPath = privateKeyPath
	cfg.PublicKeyID = "PUB_KEY_ID_test"
	cfg.PublicKeyPath = publicKeyPath
	if _, err := NewGateway(context.Background(), cfg); err != nil {
		t.Fatalf("initialize gateway with WeChat Pay public key: %v", err)
	}
}

func testPayConfig(apiV3Key, notifyURL string) config.WeChatPayConfig {
	return config.WeChatPayConfig{
		MerchantID:        "demo",
		AppID:             "app-id",
		AppSecret:         "app-secret",
		MchID:             "merchant-number",
		CertificateSerial: "serial",
		APIv3Key:          apiV3Key,
		PrivateKeyPath:    "not-loaded.pem",
		NotifyURL:         notifyURL,
	}
}
