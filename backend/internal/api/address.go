package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mendian-backend/internal/model"
	"mendian-backend/internal/payment"
)

type addressRequest struct {
	Code      string `json:"code"`
	Recipient string `json:"recipient"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail"`
	IsDefault bool   `json:"isDefault"`
}

type addressOwner struct {
	MerchantID string
	AppID      string
	OpenID     string
}

var addressPhonePattern = regexp.MustCompile(`^[0-9+() -]{6,24}$`)

func registerAddressRoutes(app *iris.Application, db *gorm.DB, gateways map[string]*payment.Gateway) {
	resolveOwner := func(ctx iris.Context, code string) (addressOwner, bool) {
		merchantID := ctx.Params().Get("merchantId")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return addressOwner{}, false
		}
		if strings.TrimSpace(code) == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "WeChat login code is required"})
			return addressOwner{}, false
		}
		openID, err := gateway.ResolveOpenID(ctx.Request().Context(), code)
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "WeChat login failed; please retry"})
			return addressOwner{}, false
		}
		if err := persistWeChatUser(db, gateway.AppID(), openID); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save WeChat user"})
			return addressOwner{}, false
		}
		return addressOwner{MerchantID: merchantID, AppID: gateway.AppID(), OpenID: openID}, true
	}

	app.Post("/api/merchants/{merchantId:string}/addresses/list", func(ctx iris.Context) {
		var request struct {
			Code string `json:"code"`
		}
		if err := ctx.ReadJSON(&request); err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		owner, ok := resolveOwner(ctx, request.Code)
		if !ok {
			return
		}
		var addresses []model.UserAddress
		if err := db.Where("merchant_id = ? AND app_id = ? AND open_id = ?", owner.MerchantID, owner.AppID, owner.OpenID).
			Order("is_default DESC, updated_at DESC, id DESC").Find(&addresses).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load addresses"})
			return
		}
		ctx.JSON(iris.Map{"addresses": addresses})
	})

	app.Post("/api/merchants/{merchantId:string}/addresses", func(ctx iris.Context) {
		request, valid := readAddressRequest(ctx)
		if !valid {
			return
		}
		owner, ok := resolveOwner(ctx, request.Code)
		if !ok {
			return
		}
		var address model.UserAddress
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := lockAddressOwner(tx, owner); err != nil {
				return err
			}
			var count int64
			if err := tx.Model(&model.UserAddress{}).Where(addressOwnerQuery(owner)).Count(&count).Error; err != nil {
				return err
			}
			address = userAddressFromRequest(owner, request)
			address.IsDefault = request.IsDefault || count == 0
			if address.IsDefault {
				if err := clearDefaultAddress(tx, owner, 0); err != nil {
					return err
				}
			}
			return tx.Create(&address).Error
		})
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save address"})
			return
		}
		ctx.StatusCode(http.StatusCreated)
		ctx.JSON(address)
	})

	app.Patch("/api/merchants/{merchantId:string}/addresses/{addressID:string}", func(ctx iris.Context) {
		request, valid := readAddressRequest(ctx)
		if !valid {
			return
		}
		owner, ok := resolveOwner(ctx, request.Code)
		if !ok {
			return
		}
		addressID, err := strconv.ParseUint(ctx.Params().Get("addressID"), 10, 64)
		if err != nil || addressID == 0 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid address ID"})
			return
		}
		var address model.UserAddress
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := lockAddressOwner(tx, owner); err != nil {
				return err
			}
			if err := tx.Where(addressOwnerQuery(owner)).First(&address, addressID).Error; err != nil {
				return err
			}
			address.Recipient = request.Recipient
			address.Phone = request.Phone
			address.Province = request.Province
			address.City = request.City
			address.District = request.District
			address.Detail = request.Detail
			if request.IsDefault {
				if err := clearDefaultAddress(tx, owner, address.ID); err != nil {
					return err
				}
				address.IsDefault = true
			}
			return tx.Save(&address).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.NotFound()
			return
		}
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to update address"})
			return
		}
		ctx.JSON(address)
	})

	app.Delete("/api/merchants/{merchantId:string}/addresses/{addressID:string}", func(ctx iris.Context) {
		var request struct {
			Code string `json:"code"`
		}
		if err := ctx.ReadJSON(&request); err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		owner, ok := resolveOwner(ctx, request.Code)
		if !ok {
			return
		}
		addressID, err := strconv.ParseUint(ctx.Params().Get("addressID"), 10, 64)
		if err != nil || addressID == 0 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid address ID"})
			return
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := lockAddressOwner(tx, owner); err != nil {
				return err
			}
			var address model.UserAddress
			if err := tx.Where(addressOwnerQuery(owner)).First(&address, addressID).Error; err != nil {
				return err
			}
			if err := tx.Delete(&address).Error; err != nil {
				return err
			}
			if address.IsDefault {
				var next model.UserAddress
				err := tx.Where(addressOwnerQuery(owner)).Order("updated_at DESC, id DESC").First(&next).Error
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
			ctx.JSON(iris.Map{"error": "failed to delete address"})
			return
		}
		ctx.StatusCode(http.StatusNoContent)
	})
}

func readAddressRequest(ctx iris.Context) (addressRequest, bool) {
	var request addressRequest
	if err := ctx.ReadJSON(&request); err != nil || strings.TrimSpace(request.Code) == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "WeChat login code is required"})
		return addressRequest{}, false
	}
	request.Recipient = strings.TrimSpace(request.Recipient)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Province = strings.TrimSpace(request.Province)
	request.City = strings.TrimSpace(request.City)
	request.District = strings.TrimSpace(request.District)
	request.Detail = strings.TrimSpace(request.Detail)
	if !validAddressFields(request) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "invalid address fields"})
		return addressRequest{}, false
	}
	return request, true
}

func validAddressFields(request addressRequest) bool {
	digitCount := 0
	for _, character := range request.Phone {
		if character >= '0' && character <= '9' {
			digitCount++
		}
	}
	return request.Recipient != "" && len([]rune(request.Recipient)) <= 64 &&
		addressPhonePattern.MatchString(request.Phone) && digitCount >= 6 &&
		request.Province != "" && len([]rune(request.Province)) <= 64 &&
		request.City != "" && len([]rune(request.City)) <= 64 &&
		request.District != "" && len([]rune(request.District)) <= 64 &&
		request.Detail != "" && len([]rune(request.Detail)) <= 255
}

func userAddressFromRequest(owner addressOwner, request addressRequest) model.UserAddress {
	return model.UserAddress{
		MerchantID: owner.MerchantID,
		AppID:      owner.AppID,
		OpenID:     owner.OpenID,
		Recipient:  request.Recipient,
		Phone:      request.Phone,
		Province:   request.Province,
		City:       request.City,
		District:   request.District,
		Detail:     request.Detail,
		IsDefault:  request.IsDefault,
	}
}

func addressOwnerQuery(owner addressOwner) map[string]any {
	return map[string]any{"merchant_id": owner.MerchantID, "app_id": owner.AppID, "open_id": owner.OpenID}
}

func clearDefaultAddress(tx *gorm.DB, owner addressOwner, exceptID uint) error {
	query := tx.Model(&model.UserAddress{}).Where(addressOwnerQuery(owner))
	if exceptID > 0 {
		query = query.Where("id <> ?", exceptID)
	}
	return query.Update("is_default", false).Error
}

func lockAddressOwner(tx *gorm.DB, owner addressOwner) error {
	var user model.WeChatUser
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("app_id = ? AND openid = ?", owner.AppID, owner.OpenID).
		First(&user).Error
}
