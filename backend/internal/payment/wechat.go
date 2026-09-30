package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"mendian-backend/internal/config"
)

const sessionURL = "https://api.weixin.qq.com/sns/jscode2session"

type PaymentParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type Gateway struct {
	config        config.WeChatPayConfig
	client        *core.Client
	notifyHandler *notify.Handler
	sessionClient *http.Client
}

// NewIdentityGateway creates the WeChat session client without initializing
// payment signing or notification verification.
func NewIdentityGateway(cfg config.WeChatPayConfig) *Gateway {
	return &Gateway{config: cfg, sessionClient: &http.Client{Timeout: 10 * time.Second}}
}

func NewGateway(ctx context.Context, cfg config.WeChatPayConfig) (*Gateway, error) {
	if !cfg.Configured() {
		return nil, errors.New("WeChat Pay is not fully configured")
	}
	if len(cfg.APIv3Key) != 32 {
		return nil, errors.New("WeChat Pay API v3 key must be 32 bytes")
	}
	callbackURL, err := url.Parse(cfg.NotifyURL)
	if err != nil || callbackURL.Scheme != "https" || callbackURL.Host == "" || callbackURL.User != nil || callbackURL.RawQuery != "" || callbackURL.ForceQuery || callbackURL.Fragment != "" {
		return nil, errors.New("WeChat Pay notify URL must be an HTTPS URL without a query string")
	}
	if callbackURL.Path != "/api/payment/wechat/notify/"+cfg.MerchantID {
		return nil, errors.New("WeChat Pay notify URL path must match the configured merchant")
	}
	host := strings.ToLower(callbackURL.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return nil, errors.New("WeChat Pay notify URL must be publicly reachable")
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified()) {
		return nil, errors.New("WeChat Pay notify URL must be publicly reachable")
	}

	privateKey, err := utils.LoadPrivateKeyWithPath(cfg.PrivateKeyPath)
	if err != nil {
		return nil, errors.New("failed to load WeChat merchant private key")
	}
	var client *core.Client
	var verifier auth.Verifier
	if cfg.PublicKeyID != "" || cfg.PublicKeyPath != "" {
		if cfg.PublicKeyID == "" || cfg.PublicKeyPath == "" {
			return nil, errors.New("WeChat Pay public key ID and path must both be configured")
		}
		publicKey, err := utils.LoadPublicKeyWithPath(cfg.PublicKeyPath)
		if err != nil {
			return nil, errors.New("failed to load WeChat Pay public key")
		}
		client, err = core.NewClient(ctx, option.WithWechatPayPublicKeyAuthCipher(
			cfg.MchID, cfg.CertificateSerial, privateKey, cfg.PublicKeyID, publicKey,
		))
		if err != nil {
			return nil, fmt.Errorf("initialize WeChat Pay client: %w", err)
		}
		verifier = verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.PublicKeyID, *publicKey)
	} else {
		client, err = core.NewClient(ctx, option.WithWechatPayAutoAuthCipher(
			cfg.MchID, cfg.CertificateSerial, privateKey, cfg.APIv3Key,
		))
		if err != nil {
			return nil, fmt.Errorf("initialize WeChat Pay client: %w", err)
		}
		certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
		verifier = verifiers.NewSHA256WithRSAVerifier(certificateVisitor)
	}
	notifyHandler, err := notify.NewRSANotifyHandler(cfg.APIv3Key, verifier)
	if err != nil {
		return nil, fmt.Errorf("initialize WeChat Pay notification handler: %w", err)
	}

	return &Gateway{
		config:        cfg,
		client:        client,
		notifyHandler: notifyHandler,
		sessionClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (g *Gateway) MerchantID() string { return g.config.MerchantID }
func (g *Gateway) AppID() string      { return g.config.AppID }
func (g *Gateway) MchID() string      { return g.config.MchID }

func (g *Gateway) ResolveOpenID(ctx context.Context, code string) (string, error) {
	if code == "" || len(code) > 512 {
		return "", errors.New("invalid WeChat login code")
	}
	query := url.Values{
		"appid":      {g.config.AppID},
		"secret":     {g.config.AppSecret},
		"js_code":    {code},
		"grant_type": {"authorization_code"},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sessionURL+"?"+query.Encode(), nil)
	if err != nil {
		return "", errors.New("failed to create WeChat login request")
	}
	response, err := g.sessionClient.Do(request)
	if err != nil {
		return "", errors.New("failed to exchange WeChat login code")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", errors.New("WeChat login service returned an error")
	}
	var result struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return "", errors.New("invalid response from WeChat login service")
	}
	if result.ErrCode != 0 || result.OpenID == "" {
		return "", errors.New("WeChat login code could not be verified")
	}
	return result.OpenID, nil
}

func (g *Gateway) CreateJSAPI(ctx context.Context, orderNo, description string, totalAmount int64, openID string) (PaymentParams, string, error) {
	expiresAt := time.Now().Add(15 * time.Minute)
	service := jsapi.JsapiApiService{Client: g.client}
	response, _, err := service.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(g.config.AppID),
		Mchid:       core.String(g.config.MchID),
		Description: core.String(description),
		OutTradeNo:  core.String(orderNo),
		TimeExpire:  &expiresAt,
		NotifyUrl:   core.String(g.config.NotifyURL),
		Amount: &jsapi.Amount{
			Total:    core.Int64(totalAmount),
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{Openid: core.String(openID)},
	})
	if err != nil {
		return PaymentParams{}, "", fmt.Errorf("failed to create WeChat prepayment: %w", err)
	}
	if response == nil || response.PrepayId == nil || response.TimeStamp == nil || response.NonceStr == nil || response.Package == nil || response.SignType == nil || response.PaySign == nil {
		return PaymentParams{}, "", errors.New("WeChat returned incomplete payment parameters")
	}
	return PaymentParams{
		TimeStamp: *response.TimeStamp,
		NonceStr:  *response.NonceStr,
		Package:   *response.Package,
		SignType:  *response.SignType,
		PaySign:   *response.PaySign,
	}, *response.PrepayId, nil
}

func (g *Gateway) ParseNotification(ctx context.Context, request *http.Request) (*payments.Transaction, error) {
	transaction := new(payments.Transaction)
	if _, err := g.notifyHandler.ParseNotifyRequest(ctx, request, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}
