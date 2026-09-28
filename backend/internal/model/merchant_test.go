package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMerchantJSONOmitsPaymentSecrets(t *testing.T) {
	merchant := Merchant{
		MerchantID:                 "merchant-1",
		WeChatAPIv3KeyCiphertext:   "api-v3-secret",
		WeChatPrivateKeyCiphertext: "private-key-secret",
	}

	data, err := json.Marshal(merchant)
	if err != nil {
		t.Fatalf("marshal merchant: %v", err)
	}

	encoded := string(data)
	if !strings.Contains(encoded, `"merchantId":"merchant-1"`) {
		t.Fatalf("merchant ID missing from JSON: %s", encoded)
	}
	if strings.Contains(encoded, "api-v3-secret") || strings.Contains(encoded, "private-key-secret") {
		t.Fatalf("payment secrets leaked in JSON: %s", encoded)
	}
}
