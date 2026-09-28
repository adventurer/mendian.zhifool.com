package api

import (
	"regexp"
	"sync"
	"testing"

	"gorm.io/gorm/schema"

	"mendian-backend/internal/model"
)

func TestNewOrderNoFitsWeChatPayFormat(t *testing.T) {
	validOrderNo := regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)
	seen := make(map[string]struct{}, 100)
	for range 100 {
		orderNo, err := newOrderNo()
		if err != nil {
			t.Fatalf("generate order number: %v", err)
		}
		if !validOrderNo.MatchString(orderNo) {
			t.Fatalf("generated invalid WeChat order number %q (length %d)", orderNo, len(orderNo))
		}
		if _, ok := seen[orderNo]; ok {
			t.Fatalf("generated duplicate order number %q", orderNo)
		}
		seen[orderNo] = struct{}{}
	}
}

func TestPaymentOrderUpdatesUseMappedColumnNames(t *testing.T) {
	orderSchema, err := schema.Parse(&model.PaymentOrder{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse payment order schema: %v", err)
	}

	for _, field := range []struct {
		name   string
		column string
	}{
		{name: "WeChatPrepayID", column: "we_chat_prepay_id"},
		{name: "WeChatTransactionID", column: "we_chat_transaction_id"},
		{name: "FulfillmentType", column: "fulfillment_type"},
	} {
		mapped := orderSchema.LookUpField(field.name)
		if mapped == nil || mapped.DBName != field.column {
			t.Fatalf("field %s maps to %v, want column %s", field.name, mapped, field.column)
		}
	}
}

func TestNormalizeOrderFulfillmentTypeDefaultsToDineIn(t *testing.T) {
	got, ok := normalizeOrderFulfillmentType("")
	if !ok || got != model.OrderFulfillmentDineIn {
		t.Fatalf("empty fulfillment type should default to dine in, got %q, valid=%t", got, ok)
	}
	for _, value := range []model.OrderFulfillmentType{model.OrderFulfillmentDineIn, model.OrderFulfillmentDelivery} {
		got, ok := normalizeOrderFulfillmentType(value)
		if !ok || got != value {
			t.Fatalf("expected %q to be valid, got %q, valid=%t", value, got, ok)
		}
	}
	if _, ok := normalizeOrderFulfillmentType("curbside"); ok {
		t.Fatal("expected unsupported fulfillment type to be rejected")
	}
}
