package model

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestUserAddressOwnershipFieldsArePrivate(t *testing.T) {
	address := UserAddress{MerchantID: "merchant", AppID: "app", OpenID: "openid", Recipient: "张女士"}
	encoded, err := json.Marshal(address)
	if err != nil {
		t.Fatalf("marshal address: %v", err)
	}
	if strings.Contains(string(encoded), "openid") || strings.Contains(string(encoded), "merchant") || strings.Contains(string(encoded), "app") {
		t.Fatalf("address ownership leaked in JSON: %s", encoded)
	}
	parsed, err := schema.Parse(&UserAddress{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse address schema: %v", err)
	}
	if parsed.Table != "user_addresses" {
		t.Fatalf("unexpected address table %q", parsed.Table)
	}
	openIDField := parsed.LookUpField("OpenID")
	if openIDField == nil || openIDField.DBName != "open_id" {
		t.Fatalf("OpenID must map to open_id, got %v", openIDField)
	}
}
