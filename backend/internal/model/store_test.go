package model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestStoreUsesMerchantScopedIdentity(t *testing.T) {
	parsed, err := schema.Parse(&Store{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse store schema: %v", err)
	}
	if parsed.Table != "merchant_stores" {
		t.Fatalf("unexpected store table name %q", parsed.Table)
	}
	for _, fieldName := range []string{"MerchantID", "StoreID"} {
		field := parsed.LookUpField(fieldName)
		if field == nil || !field.PrimaryKey {
			t.Fatalf("expected %s to be part of the store primary key", fieldName)
		}
	}
}

func TestCartAndOrderModelsPersistStoreIdentity(t *testing.T) {
	for _, entry := range []struct {
		model  any
		fields map[string]string
	}{
		{model: &CartItem{}, fields: map[string]string{"StoreID": "store_id"}},
		{model: &PaymentOrder{}, fields: map[string]string{"StoreID": "store_id", "StoreName": "store_name"}},
	} {
		parsed, err := schema.Parse(entry.model, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatalf("parse store-scoped model schema: %v", err)
		}
		for name, column := range entry.fields {
			field := parsed.LookUpField(name)
			if field == nil || field.DBName != column {
				t.Fatalf("field %s maps to %v, want column %s", name, field, column)
			}
		}
	}
}

func TestStoreCatalogSchemasParse(t *testing.T) {
	for _, candidate := range []any{
		&StoreMenu{},
		&StoreMenuSeries{},
		&StoreProduct{},
		&StoreProductOption{},
		&StoreProductOptionValue{},
	} {
		if _, err := schema.Parse(candidate, &sync.Map{}, schema.NamingStrategy{}); err != nil {
			t.Fatalf("parse store catalog schema %T: %v", candidate, err)
		}
	}
}
