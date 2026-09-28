package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config.yaml"

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	WeChatPay WeChatPayConfig `yaml:"wechatPay"`
}

type WeChatPayConfig struct {
	MerchantID        string                     `yaml:"merchantId"`
	AppID             string                     `yaml:"appId"`
	AppSecret         string                     `yaml:"appSecret"`
	MchID             string                     `yaml:"mchId"`
	CertificateSerial string                     `yaml:"certificateSerial"`
	PublicKeyID       string                     `yaml:"publicKeyId"`
	PublicKeyPath     string                     `yaml:"publicKeyPath"`
	APIv3Key          string                     `yaml:"apiV3Key"`
	PrivateKeyPath    string                     `yaml:"privateKeyPath"`
	NotifyURL         string                     `yaml:"notifyUrl"`
	Merchants         map[string]WeChatPayConfig `yaml:"merchants"`
}

func (cfg WeChatPayConfig) Configured() bool {
	return cfg.MerchantID != "" && cfg.AppID != "" && cfg.AppSecret != "" &&
		cfg.MchID != "" && cfg.CertificateSerial != "" && cfg.APIv3Key != "" &&
		cfg.PrivateKeyPath != "" && cfg.NotifyURL != ""
}

func (cfg WeChatPayConfig) MerchantConfigs() []WeChatPayConfig {
	if len(cfg.Merchants) > 0 {
		merchantIDs := make([]string, 0, len(cfg.Merchants))
		for merchantID := range cfg.Merchants {
			merchantIDs = append(merchantIDs, merchantID)
		}
		sort.Strings(merchantIDs)
		configs := make([]WeChatPayConfig, 0, len(merchantIDs))
		for _, merchantID := range merchantIDs {
			merchantConfig := cfg.Merchants[merchantID]
			merchantConfig.MerchantID = merchantID
			configs = append(configs, merchantConfig)
		}
		return configs
	}
	if cfg.Configured() {
		return []WeChatPayConfig{cfg}
	}
	return nil
}

type ServerConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	TLSCert       string `yaml:"tlsCert"`
	TLSKey        string `yaml:"tlsKey"`
	AutoTLSDomain string `yaml:"autoTlsDomain"`
	AutoTLSEmail  string `yaml:"autoTlsEmail"`
}

type DatabaseConfig struct {
	Driver      string `yaml:"driver"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	PasswordEnv string `yaml:"passwordEnv"`
	Name        string `yaml:"name"`
	Charset     string `yaml:"charset"`
	ParseTime   bool   `yaml:"parseTime"`
	Location    string `yaml:"location"`
}

func Load(path string) (Config, error) {
	if path == "" {
		path = defaultConfigPath
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(contents, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}
	if err := applyEnvironment(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyEnvironment(cfg *Config) error {
	stringOverrides := []struct {
		name  string
		value *string
	}{
		{"SERVER_HOST", &cfg.Server.Host},
		{"SERVER_AUTOTLS_DOMAIN", &cfg.Server.AutoTLSDomain},
		{"SERVER_AUTOTLS_EMAIL", &cfg.Server.AutoTLSEmail},
		{"DB_HOST", &cfg.Database.Host},
		{"DB_USERNAME", &cfg.Database.Username},
		{"DB_NAME", &cfg.Database.Name},
		{"DB_CHARSET", &cfg.Database.Charset},
		{"WECHAT_PAY_MERCHANT_ID", &cfg.WeChatPay.MerchantID},
		{"WECHAT_APP_ID", &cfg.WeChatPay.AppID},
		{"WECHAT_APP_SECRET", &cfg.WeChatPay.AppSecret},
		{"WECHAT_MCH_ID", &cfg.WeChatPay.MchID},
		{"WECHAT_MCH_CERTIFICATE_SERIAL", &cfg.WeChatPay.CertificateSerial},
		{"WECHAT_PAY_PUBLIC_KEY_ID", &cfg.WeChatPay.PublicKeyID},
		{"WECHAT_PAY_PUBLIC_KEY_PATH", &cfg.WeChatPay.PublicKeyPath},
		{"WECHAT_PAY_API_V3_KEY", &cfg.WeChatPay.APIv3Key},
		{"WECHAT_MCH_PRIVATE_KEY_PATH", &cfg.WeChatPay.PrivateKeyPath},
		{"WECHAT_PAY_NOTIFY_URL", &cfg.WeChatPay.NotifyURL},
	}
	for _, override := range stringOverrides {
		if value, ok := os.LookupEnv(override.name); ok {
			*override.value = value
		}
	}

	intOverrides := []struct {
		name  string
		value *int
	}{
		{"SERVER_PORT", &cfg.Server.Port},
		{"DB_PORT", &cfg.Database.Port},
	}
	for _, override := range intOverrides {
		if value, ok := os.LookupEnv(override.name); ok {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("parse %s: %w", override.name, err)
			}
			*override.value = parsed
		}
	}

	passwordEnv := cfg.Database.PasswordEnv
	if passwordEnv == "" {
		passwordEnv = "DB_PASSWORD"
	}
	if value := os.Getenv(passwordEnv); value != "" {
		cfg.Database.Password = value
	}
	if value := os.Getenv("DB_PASSWORD"); value != "" {
		cfg.Database.Password = value
	}

	return nil
}

func (cfg Config) validate() error {
	if cfg.Server.Host == "" || cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server host or port")
	}
	if (cfg.Server.TLSCert == "") != (cfg.Server.TLSKey == "") {
		return fmt.Errorf("both server TLS certificate and key must be configured")
	}
	if cfg.Server.AutoTLSDomain != "" && cfg.Server.TLSCert != "" {
		return fmt.Errorf("automatic TLS and manual TLS certificates cannot both be configured")
	}
	if cfg.Database.Driver != "mysql" {
		return fmt.Errorf("unsupported database driver %q", cfg.Database.Driver)
	}
	if cfg.Database.Host == "" || cfg.Database.Username == "" || cfg.Database.Name == "" || cfg.Database.Charset == "" {
		return fmt.Errorf("database host, username, name, and charset are required")
	}
	if cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return fmt.Errorf("invalid database port %d", cfg.Database.Port)
	}
	if len(cfg.WeChatPay.Merchants) > 0 {
		for merchantID, merchant := range cfg.WeChatPay.Merchants {
			if merchantID == "" {
				return fmt.Errorf("WeChat Pay merchant ID key cannot be empty")
			}
			if merchant.MerchantID != "" && merchant.MerchantID != merchantID {
				return fmt.Errorf("WeChat Pay merchant ID %q does not match configuration key %q", merchant.MerchantID, merchantID)
			}
			merchant.MerchantID = merchantID
			if !merchant.Configured() {
				return fmt.Errorf("incomplete WeChat Pay configuration for merchant %q", merchant.MerchantID)
			}
		}
	}
	return nil
}
