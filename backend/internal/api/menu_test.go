package api

import (
	"math"
	"testing"

	"mendian-backend/internal/model"
	"mendian-backend/internal/seed"
)

func TestBuildCartItemCalculatesPriceFromMenu(t *testing.T) {
	menu := seed.DefaultMenu()
	item, err := buildCartItem(menu.MerchantID, "cart-test", menu.Series, cartItemRequest{
		ProductID: "rose-rice-latte",
		Quantity:  2,
		Options: []optionSelection{
			{ID: "size", SelectedIndex: 1},
			{ID: "milk", SelectedIndex: 1},
			{ID: "sweetness", SelectedIndex: 2},
		},
	})
	if err != nil {
		t.Fatalf("build cart item: %v", err)
	}
	if item.MerchantID != seed.DefaultMerchantID || item.CartID != "cart-test" || item.UnitPrice != 28 || item.Quantity != 2 {
		t.Fatalf("unexpected cart item: %+v", item)
	}
	if item.Options != "超大热杯 473ml，燕麦奶，不另外加糖" {
		t.Fatalf("unexpected options summary: %q", item.Options)
	}
}

func TestSnapshotCartUsesDatabasePricesAndConvertsToFen(t *testing.T) {
	snapshot, err := snapshotCart([]model.CartItem{
		{ID: 7, ProductID: "latte", Name: "拿铁", Quantity: 2, UnitPrice: 20.25, Options: "大杯"},
		{ID: 8, ProductID: "tea", Name: "茶咖", Quantity: 1, UnitPrice: 3.5},
	})
	if err != nil {
		t.Fatalf("snapshot cart: %v", err)
	}
	if snapshot.Total != 4400 || len(snapshot.Items) != 2 {
		t.Fatalf("unexpected order total/items: %+v", snapshot)
	}
	if snapshot.Items[0].CartItemID != 7 || snapshot.Items[0].UnitAmount != 2025 || snapshot.Items[0].LineAmount != 4050 {
		t.Fatalf("unexpected item snapshot: %+v", snapshot.Items[0])
	}
}

func TestSnapshotCartRejectsInvalidCartAndPrices(t *testing.T) {
	if _, err := snapshotCart(nil); err != errEmptyCart {
		t.Fatalf("expected empty-cart error, got %v", err)
	}
	invalidCarts := [][]model.CartItem{
		{{ID: 1, Quantity: 0, UnitPrice: 1}},
		{{ID: 1, Quantity: 1, UnitPrice: math.NaN()}},
		{{ID: 1, Quantity: 1, UnitPrice: -1}},
		{{ID: 1, Quantity: 1, UnitPrice: float64(maxInt64) / 100}},
	}
	for _, cart := range invalidCarts {
		if _, err := snapshotCart(cart); err != errInvalidCartTotal {
			t.Fatalf("expected invalid-total error, got %v", err)
		}
	}
}

func TestBuildCartItemRejectsInvalidSelection(t *testing.T) {
	menu := seed.DefaultMenu()
	_, err := buildCartItem(menu.MerchantID, "cart-test", menu.Series, cartItemRequest{
		ProductID: "rose-rice-latte",
		Quantity:  1,
		Options:   []optionSelection{{ID: "size", SelectedIndex: 100}},
	})
	if err == nil {
		t.Fatal("expected invalid option selection to fail")
	}
}

func TestStoreMenuMapsStoreProductsAndOptionsToPublicMenu(t *testing.T) {
	menu := storeMenuToMenu(model.StoreMenu{
		MerchantID: "merchant",
		StoreID:    "branch",
		BrandName:  "Brand",
		Series: []model.StoreMenuSeries{{
			ID:   "drinks",
			Name: "饮品",
			Products: []model.StoreProduct{{
				ID:       "latte",
				SeriesID: "drinks",
				Name:     "拿铁",
				Price:    18,
				Options: []model.StoreProductOption{{
					ID:     "size",
					Title:  "杯型",
					Values: []model.StoreProductOptionValue{{Label: "大杯", ExtraPrice: 2}},
				}},
			}},
		}},
	})
	if menu.MerchantID != "merchant" || len(menu.Series) != 1 || len(menu.Series[0].Products) != 1 {
		t.Fatalf("unexpected mapped store menu: %+v", menu)
	}
	product := menu.Series[0].Products[0]
	if product.ID != "latte" || product.Price != 18 || len(product.Options) != 1 || product.Options[0].Values[0].ExtraPrice != 2 {
		t.Fatalf("store catalog product/options were not mapped: %+v", product)
	}
}
