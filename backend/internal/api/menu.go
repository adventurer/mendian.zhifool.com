package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"mendian-backend/internal/model"
)

type cartItemRequest struct {
	ProductID string            `json:"productId"`
	Quantity  int               `json:"quantity"`
	Options   []optionSelection `json:"options"`
}

type optionSelection struct {
	ID            string `json:"id"`
	SelectedIndex int    `json:"selectedIndex"`
}

type cartItemQuantityRequest struct {
	Quantity int `json:"quantity"`
}

func RegisterRoutes(app *iris.Application, db *gorm.DB) {
	app.Get("/api/merchants/{merchantId:string}/stores", func(ctx iris.Context) {
		var stores []model.Store
		if err := db.Table("merchant_stores AS stores").
			Select("stores.*").
			Joins("JOIN store_menus ON store_menus.merchant_id = stores.merchant_id AND store_menus.store_id = stores.store_id").
			Where("stores.merchant_id = ? AND stores.is_active = ?", ctx.Params().Get("merchantId"), true).
			Order("stores.is_default DESC, stores.name ASC").Scan(&stores).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load stores"})
			return
		}
		ctx.JSON(stores)
	})

	app.Get("/api/merchants/{merchantId:string}/menu", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		menu, err := findMenu(db, merchantID, store.StoreID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.NotFound()
				return
			}
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load menu"})
			return
		}
		ctx.JSON(&menu)
	})

	app.Get("/api/merchants/{merchantId:string}/carts/{cartId:string}/items", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		var items []model.CartItem
		if err := db.Where("merchant_id = ? AND store_id = ? AND cart_id = ?", merchantID, store.StoreID, ctx.Params().Get("cartId")).Order("id ASC").Find(&items).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load cart"})
			return
		}
		ctx.JSON(items)
	})

	app.Patch("/api/merchants/{merchantId:string}/carts/{cartId:string}/items/{itemID:string}", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		itemID, err := strconv.ParseUint(ctx.Params().Get("itemID"), 10, 64)
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid cart item ID"})
			return
		}
		var request cartItemQuantityRequest
		if err := ctx.ReadJSON(&request); err != nil || request.Quantity < 1 || request.Quantity > 99 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "quantity must be between 1 and 99"})
			return
		}

		result := db.Model(&model.CartItem{}).
			Where("merchant_id = ? AND store_id = ? AND cart_id = ? AND id = ?", merchantID, store.StoreID, ctx.Params().Get("cartId"), itemID).
			Update("quantity", request.Quantity)
		if result.Error != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to update cart item"})
			return
		}
		if result.RowsAffected == 0 {
			ctx.NotFound()
			return
		}

		var item model.CartItem
		if err := db.Where("merchant_id = ? AND store_id = ? AND cart_id = ? AND id = ?", merchantID, store.StoreID, ctx.Params().Get("cartId"), itemID).First(&item).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load cart item"})
			return
		}
		ctx.JSON(item)
	})

	app.Delete("/api/merchants/{merchantId:string}/carts/{cartId:string}/items/{itemID:string}", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		itemID, err := strconv.ParseUint(ctx.Params().Get("itemID"), 10, 64)
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid cart item ID"})
			return
		}
		result := db.Where("merchant_id = ? AND store_id = ? AND cart_id = ? AND id = ?", merchantID, store.StoreID, ctx.Params().Get("cartId"), itemID).Delete(&model.CartItem{})
		if result.Error != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to remove cart item"})
			return
		}
		if result.RowsAffected == 0 {
			ctx.NotFound()
			return
		}
		ctx.StatusCode(http.StatusNoContent)
	})

	app.Post("/api/merchants/{merchantId:string}/carts/{cartId:string}/items", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		var request cartItemRequest
		if err := ctx.ReadJSON(&request); err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}

		menu, err := findMenu(db, merchantID, store.StoreID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.NotFound()
				return
			}
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load menu"})
			return
		}

		createdItem, err := buildCartItem(menu.MerchantID, ctx.Params().Get("cartId"), menu.Series, request)
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": err.Error()})
			return
		}
		createdItem.StoreID = store.StoreID
		if err := db.Create(&createdItem).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save cart item"})
			return
		}

		ctx.StatusCode(http.StatusCreated)
		ctx.JSON(createdItem)
	})
}

func requireActiveStore(ctx iris.Context, db *gorm.DB, merchantID string) (model.Store, bool) {
	storeID := strings.TrimSpace(ctx.URLParam("storeId"))
	if storeID == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "store ID is required"})
		return model.Store{}, false
	}
	var store model.Store
	err := db.Where("merchant_id = ? AND store_id = ? AND is_active = ?", merchantID, storeID, true).First(&store).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.NotFound()
		return model.Store{}, false
	}
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "failed to load store"})
		return model.Store{}, false
	}
	return store, true
}

func findMenu(db *gorm.DB, merchantID, storeID string) (model.Menu, error) {
	var storeMenu model.StoreMenu
	ordered := func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }
	err := db.
		Preload("Series", ordered).
		Preload("Series.Products", ordered).
		Preload("Series.Products.Options", ordered).
		Preload("Series.Products.Options.Values", ordered).
		Where("merchant_id = ? AND store_id = ?", merchantID, storeID).
		First(&storeMenu).Error
	if err != nil {
		return model.Menu{}, err
	}
	return storeMenuToMenu(storeMenu), nil
}

func storeMenuToMenu(storeMenu model.StoreMenu) model.Menu {
	menu := model.Menu{MerchantID: storeMenu.MerchantID, BrandName: storeMenu.BrandName, Series: make([]model.MenuSeries, 0, len(storeMenu.Series))}
	for _, storeSeries := range storeMenu.Series {
		series := model.MenuSeries{
			MerchantID: storeSeries.MerchantID,
			ID:         storeSeries.ID,
			Name:       storeSeries.Name,
			Icon:       storeSeries.Icon,
			Badge:      storeSeries.Badge,
			Products:   make([]model.Product, 0, len(storeSeries.Products)),
		}
		for _, storeProduct := range storeSeries.Products {
			product := model.Product{
				MerchantID:  storeProduct.MerchantID,
				ID:          storeProduct.ID,
				SeriesID:    storeProduct.SeriesID,
				Name:        storeProduct.Name,
				Description: storeProduct.Description,
				Price:       storeProduct.Price,
				Image:       storeProduct.Image,
				Options:     make([]model.ProductOption, 0, len(storeProduct.Options)),
			}
			for _, storeOption := range storeProduct.Options {
				option := model.ProductOption{
					MerchantID:    storeOption.MerchantID,
					ProductID:     storeOption.ProductID,
					ID:            storeOption.ID,
					Title:         storeOption.Title,
					SelectedIndex: storeOption.SelectedIndex,
					Values:        make([]model.ProductOptionValue, 0, len(storeOption.Values)),
				}
				for _, storeValue := range storeOption.Values {
					option.Values = append(option.Values, model.ProductOptionValue{ID: storeValue.ID, Label: storeValue.Label, ExtraPrice: storeValue.ExtraPrice})
				}
				product.Options = append(product.Options, option)
			}
			series.Products = append(series.Products, product)
		}
		menu.Series = append(menu.Series, series)
	}
	return menu
}

func buildCartItem(merchantID, cartID string, series []model.MenuSeries, request cartItemRequest) (model.CartItem, error) {
	if cartID == "" || len(cartID) > 64 {
		return model.CartItem{}, errors.New("invalid cart ID")
	}
	if request.Quantity < 1 || request.Quantity > 99 {
		return model.CartItem{}, errors.New("quantity must be between 1 and 99")
	}

	var selectedProduct *model.Product
	for seriesIndex := range series {
		for productIndex := range series[seriesIndex].Products {
			if series[seriesIndex].Products[productIndex].ID == request.ProductID {
				selectedProduct = &series[seriesIndex].Products[productIndex]
				break
			}
		}
		if selectedProduct != nil {
			break
		}
	}
	if selectedProduct == nil {
		return model.CartItem{}, errors.New("product not found")
	}

	selections := make(map[string]int, len(request.Options))
	for _, option := range request.Options {
		if _, exists := selections[option.ID]; exists {
			return model.CartItem{}, errors.New("duplicate product option")
		}
		selections[option.ID] = option.SelectedIndex
	}

	price := selectedProduct.Price
	selectedLabels := make([]string, 0, len(selectedProduct.Options))
	for _, group := range selectedProduct.Options {
		index := group.SelectedIndex
		if selectedIndex, exists := selections[group.ID]; exists {
			index = selectedIndex
		}
		if index < 0 || index >= len(group.Values) {
			return model.CartItem{}, errors.New("invalid product option")
		}
		value := group.Values[index]
		price += value.ExtraPrice
		selectedLabels = append(selectedLabels, value.Label)
		delete(selections, group.ID)
	}
	if len(selections) > 0 {
		return model.CartItem{}, errors.New("unknown product option")
	}
	if price < 0 {
		return model.CartItem{}, errors.New("invalid product price")
	}

	return model.CartItem{
		MerchantID: merchantID,
		CartID:     cartID,
		ProductID:  selectedProduct.ID,
		Name:       selectedProduct.Name,
		Quantity:   request.Quantity,
		UnitPrice:  price,
		Options:    strings.Join(selectedLabels, "，"),
	}, nil
}
