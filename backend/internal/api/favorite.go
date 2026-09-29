package api

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"mendian-backend/internal/model"
	"mendian-backend/internal/payment"
)

type favoriteRequest struct {
	Code      string `json:"code"`
	StoreID   string `json:"storeId"`
	ProductID string `json:"productId"`
}

type favoriteProduct struct {
	Product    model.Product `json:"product"`
	SeriesID   string        `json:"seriesId"`
	SeriesName string        `json:"seriesName"`
	CreatedAt  string        `json:"createdAt"`
}

func RegisterFavoriteRoutes(app *iris.Application, db *gorm.DB, gateways map[string]*payment.Gateway) {
	app.Post("/api/merchants/{merchantId:string}/favorites/list", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return
		}
		var request favoriteRequest
		if err := ctx.ReadJSON(&request); err != nil || request.Code == "" || request.StoreID == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "login code and store ID are required"})
			return
		}
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		if store.StoreID != request.StoreID {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "store ID does not match request"})
			return
		}
		openID, err := gateway.ResolveOpenID(ctx.Request().Context(), request.Code)
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "WeChat login failed; please retry"})
			return
		}
		if err := persistWeChatUser(db, gateway.AppID(), openID); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save WeChat user"})
			return
		}
		var favorites []model.UserFavorite
		if err := db.Where("merchant_id = ? AND store_id = ? AND app_id = ? AND open_id = ?", merchantID, store.StoreID, gateway.AppID(), openID).
			Order("created_at DESC, product_id ASC").Find(&favorites).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load favorites"})
			return
		}
		menu, err := findMenu(db, merchantID, store.StoreID)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load store menu"})
			return
		}
		products := make(map[string]favoriteProduct)
		for _, series := range menu.Series {
			for _, product := range series.Products {
				products[product.ID] = favoriteProduct{Product: product, SeriesID: series.ID, SeriesName: series.Name}
			}
		}
		result := make([]favoriteProduct, 0, len(favorites))
		for _, favorite := range favorites {
			if product, exists := products[favorite.ProductID]; exists {
				product.CreatedAt = favorite.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
				result = append(result, product)
			}
		}
		ctx.JSON(iris.Map{"favorites": result})
	})

	app.Post("/api/merchants/{merchantId:string}/favorites/toggle", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return
		}
		var request favoriteRequest
		if err := ctx.ReadJSON(&request); err != nil || request.Code == "" || request.StoreID == "" || request.ProductID == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "login code, store ID and product ID are required"})
			return
		}
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		if store.StoreID != request.StoreID {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "store ID does not match request"})
			return
		}
		var productCount int64
		if err := db.Model(&model.StoreProduct{}).Where("merchant_id = ? AND store_id = ? AND id = ?", merchantID, store.StoreID, request.ProductID).Count(&productCount).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to check product"})
			return
		}
		if productCount == 0 {
			ctx.NotFound()
			return
		}
		openID, err := gateway.ResolveOpenID(ctx.Request().Context(), request.Code)
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "WeChat login failed; please retry"})
			return
		}
		if err := persistWeChatUser(db, gateway.AppID(), openID); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save WeChat user"})
			return
		}
		favorite := model.UserFavorite{MerchantID: merchantID, StoreID: store.StoreID, AppID: gateway.AppID(), OpenID: openID, ProductID: request.ProductID}
		err = db.Transaction(func(tx *gorm.DB) error {
			result := tx.Where("merchant_id = ? AND store_id = ? AND app_id = ? AND open_id = ? AND product_id = ?", favorite.MerchantID, favorite.StoreID, favorite.AppID, favorite.OpenID, favorite.ProductID).Delete(&model.UserFavorite{})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				return nil
			}
			return tx.Create(&favorite).Error
		})
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to update favorite"})
			return
		}
		var count int64
		if err := db.Model(&model.UserFavorite{}).Where("merchant_id = ? AND store_id = ? AND app_id = ? AND open_id = ? AND product_id = ?", favorite.MerchantID, favorite.StoreID, favorite.AppID, favorite.OpenID, favorite.ProductID).Count(&count).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to read favorite status"})
			return
		}
		ctx.JSON(iris.Map{"isFavorite": count > 0})
	})
}
