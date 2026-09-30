package api

import (
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mendian-backend/internal/model"
	"mendian-backend/internal/seed"
)

type merchantProvisionRequest struct {
	Name string `json:"name"`
}

type storeWriteRequest struct {
	StoreID   string  `json:"storeId"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	IsActive  bool    `json:"isActive"`
	IsDefault bool    `json:"isDefault"`
}

func RegisterMerchantAdminRoutes(app *iris.Application, db *gorm.DB) {
	app.Post("/api/admin/merchants/{merchantId:string}", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		merchantID := strings.TrimSpace(ctx.Params().Get("merchantId"))
		var req merchantProvisionRequest
		if !validScopeID(merchantID) || ctx.ReadJSON(&req) != nil || strings.TrimSpace(req.Name) == "" || len([]rune(strings.TrimSpace(req.Name))) > 128 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "valid merchant ID and name are required"})
			return
		}
		merchant := model.Merchant{MerchantID: merchantID, Name: strings.TrimSpace(req.Name)}
		if err := db.Create(&merchant).Error; err != nil {
			ctx.StatusCode(http.StatusConflict)
			ctx.JSON(iris.Map{"error": "merchant already exists or could not be created"})
			return
		}
		ctx.StatusCode(http.StatusCreated)
		ctx.JSON(iris.Map{"merchantId": merchant.MerchantID, "name": merchant.Name})
	})

	app.Get("/api/admin/merchants/{merchantId:string}/stores", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		var stores []model.Store
		if err := db.Where("merchant_id = ?", ctx.Params().Get("merchantId")).Order("is_default DESC, created_at ASC, store_id ASC").Find(&stores).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load stores"})
			return
		}
		ctx.JSON(iris.Map{"stores": stores})
	})

	app.Post("/api/admin/merchants/{merchantId:string}/stores", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		var req storeWriteRequest
		if ctx.ReadJSON(&req) != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid store request"})
			return
		}
		store, status, message := saveMerchantStore(db, ctx.Params().Get("merchantId"), strings.TrimSpace(req.StoreID), req, true)
		if message != "" {
			ctx.StatusCode(status)
			ctx.JSON(iris.Map{"error": message})
			return
		}
		ctx.StatusCode(http.StatusCreated)
		ctx.JSON(store)
	})

	app.Put("/api/admin/merchants/{merchantId:string}/stores/{storeId:string}", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		var req storeWriteRequest
		if ctx.ReadJSON(&req) != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid store request"})
			return
		}
		store, status, message := saveMerchantStore(db, ctx.Params().Get("merchantId"), strings.TrimSpace(ctx.Params().Get("storeId")), req, false)
		if message != "" {
			ctx.StatusCode(status)
			ctx.JSON(iris.Map{"error": message})
			return
		}
		ctx.JSON(store)
	})

	app.Delete("/api/admin/merchants/{merchantId:string}/stores/{storeId:string}", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		merchantID, storeID := strings.TrimSpace(ctx.Params().Get("merchantId")), strings.TrimSpace(ctx.Params().Get("storeId"))
		if !validScopeID(merchantID) || !validScopeID(storeID) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid merchant or store ID"})
			return
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := lockMerchant(tx, merchantID); err != nil {
				return err
			}
			var store model.Store
			if err := tx.Where("merchant_id = ? AND store_id = ?", merchantID, storeID).First(&store).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Store{}).Where("merchant_id = ? AND store_id = ?", merchantID, storeID).Updates(map[string]any{"is_active": false, "is_default": false}).Error; err != nil {
				return err
			}
			if store.IsDefault {
				var next model.Store
				err := tx.Where("merchant_id = ? AND is_active = ? AND store_id <> ?", merchantID, true, storeID).Order("created_at ASC, store_id ASC").First(&next).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				if err != nil {
					return err
				}
				return tx.Model(&next).Update("is_default", true).Error
			}
			return nil
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.NotFound()
			return
		}
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to deactivate store"})
			return
		}
		ctx.StatusCode(http.StatusNoContent)
	})

	app.Put("/api/admin/merchants/{merchantId:string}/stores/{storeId:string}/menu", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		merchantID, storeID := strings.TrimSpace(ctx.Params().Get("merchantId")), strings.TrimSpace(ctx.Params().Get("storeId"))
		if !validScopeID(merchantID) || !validScopeID(storeID) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid merchant or store ID"})
			return
		}
		var menu model.Menu
		if err := ctx.ReadJSON(&menu); err != nil || !validStoreMenu(menu) {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid store menu"})
			return
		}
		menu.MerchantID = merchantID
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := lockMerchant(tx, merchantID); err != nil {
				return err
			}
			var store model.Store
			if err := tx.Where("merchant_id = ? AND store_id = ?", merchantID, storeID).First(&store).Error; err != nil {
				return err
			}
			return seed.ReplaceStoreMenu(tx, storeID, menu)
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.NotFound()
			return
		}
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to replace store menu"})
			return
		}
		ctx.JSON(iris.Map{"merchantId": merchantID, "storeId": storeID, "brandName": menu.BrandName})
	})
}

func saveMerchantStore(db *gorm.DB, merchantID, storeID string, req storeWriteRequest, create bool) (model.Store, int, string) {
	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Address = strings.TrimSpace(req.Address)
	if !validScopeID(merchantID) || !validScopeID(storeID) || req.Name == "" || len([]rune(req.Name)) > 128 || len(req.Phone) > 24 || len([]rune(req.Address)) > 255 ||
		math.IsNaN(req.Latitude) || math.IsInf(req.Latitude, 0) || req.Latitude < -90 || req.Latitude > 90 ||
		math.IsNaN(req.Longitude) || math.IsInf(req.Longitude, 0) || req.Longitude < -180 || req.Longitude > 180 || (!req.IsActive && req.IsDefault) {
		return model.Store{}, http.StatusBadRequest, "invalid store fields"
	}
	var result model.Store
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := lockMerchant(tx, merchantID); err != nil {
			return err
		}
		var existing model.Store
		err := tx.Where("merchant_id = ? AND store_id = ?", merchantID, storeID).First(&existing).Error
		if create && err == nil {
			return gorm.ErrDuplicatedKey
		}
		if !create && errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !req.IsDefault {
			var otherDefaults int64
			if err := tx.Model(&model.Store{}).Where("merchant_id = ? AND store_id <> ? AND is_active = ? AND is_default = ?", merchantID, storeID, true, true).Count(&otherDefaults).Error; err != nil {
				return err
			}
			req.IsDefault = req.IsActive && otherDefaults == 0
		}
		result = model.Store{MerchantID: merchantID, StoreID: storeID, Name: req.Name, Phone: req.Phone, Address: req.Address, Latitude: req.Latitude, Longitude: req.Longitude, IsActive: req.IsActive, IsDefault: req.IsDefault}
		if req.IsDefault {
			if err := tx.Model(&model.Store{}).Where("merchant_id = ? AND store_id <> ?", merchantID, storeID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if create {
			return tx.Create(&result).Error
		}
		if err := tx.Model(&existing).Updates(map[string]any{
			"name": result.Name, "phone": result.Phone, "address": result.Address,
			"latitude": result.Latitude, "longitude": result.Longitude,
			"is_active": result.IsActive, "is_default": result.IsDefault,
		}).Error; err != nil {
			return err
		}
		if existing.IsDefault && !result.IsDefault && !result.IsActive {
			var next model.Store
			err := tx.Where("merchant_id = ? AND is_active = ? AND store_id <> ?", merchantID, true, storeID).Order("created_at ASC, store_id ASC").First(&next).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			return tx.Model(&next).Update("is_default", true).Error
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Store{}, http.StatusNotFound, "merchant or store not found"
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return model.Store{}, http.StatusConflict, "store already exists"
	}
	if err != nil {
		return model.Store{}, http.StatusInternalServerError, "failed to save store"
	}
	if !create {
		if err := db.Where("merchant_id = ? AND store_id = ?", merchantID, storeID).First(&result).Error; err != nil {
			return model.Store{}, http.StatusInternalServerError, "failed to reload store"
		}
	}
	return result, 0, ""
}

func lockMerchant(tx *gorm.DB, merchantID string) error {
	var merchant model.Merchant
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchant_id = ?", merchantID).First(&merchant).Error
}

func validScopeID(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}

func validStoreMenu(menu model.Menu) bool {
	menu.BrandName = strings.TrimSpace(menu.BrandName)
	if menu.BrandName == "" || len([]rune(menu.BrandName)) > 128 || len(menu.Series) == 0 {
		return false
	}
	seriesIDs, productIDs := map[string]bool{}, map[string]bool{}
	for _, series := range menu.Series {
		if !validScopeID(series.ID) || strings.TrimSpace(series.Name) == "" || len([]rune(series.Name)) > 128 || seriesIDs[series.ID] {
			return false
		}
		seriesIDs[series.ID] = true
		for _, product := range series.Products {
			if !validScopeID(product.ID) || strings.TrimSpace(product.Name) == "" || len([]rune(product.Name)) > 255 || productIDs[product.ID] ||
				math.IsNaN(product.Price) || math.IsInf(product.Price, 0) || product.Price < 0 {
				return false
			}
			productIDs[product.ID] = true
			optionIDs := map[string]bool{}
			for _, option := range product.Options {
				if !validScopeID(option.ID) || strings.TrimSpace(option.Title) == "" || len(option.Values) == 0 || optionIDs[option.ID] || option.SelectedIndex < 0 || option.SelectedIndex >= len(option.Values) {
					return false
				}
				optionIDs[option.ID] = true
				for _, value := range option.Values {
					if strings.TrimSpace(value.Label) == "" || math.IsNaN(value.ExtraPrice) || math.IsInf(value.ExtraPrice, 0) || value.ExtraPrice < 0 {
						return false
					}
				}
			}
		}
	}
	return len(productIDs) > 0
}
