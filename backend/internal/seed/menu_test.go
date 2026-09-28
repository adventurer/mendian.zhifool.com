package seed

import (
	"strings"
	"testing"
)

func TestDemoStoreMenuContainsOnlyNonPayableSampleProducts(t *testing.T) {
	menu := DemoStoreMenu()
	if len(menu.Series) != 1 || len(menu.Series[0].Products) != 3 {
		t.Fatalf("expected one sample series with three products, got %+v", menu.Series)
	}
	for _, product := range menu.Series[0].Products {
		if product.Price != 0 || !strings.Contains(product.Description, "测试") {
			t.Fatalf("demo product must be visibly marked and zero-priced: %+v", product)
		}
	}
}
