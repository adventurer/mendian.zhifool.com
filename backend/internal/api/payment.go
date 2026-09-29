package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mendian-backend/internal/model"
	"mendian-backend/internal/payment"
)

type createPaymentOrderRequest struct {
	Code            string                     `json:"code"`
	FulfillmentType model.OrderFulfillmentType `json:"fulfillmentType"`
	BenefitID       *uint                      `json:"benefitId"`
}

type wechatLoginRequest struct {
	MerchantID string `json:"merchantId"`
	Code       string `json:"code"`
}

type myPaymentOrdersRequest struct {
	Code   string                 `json:"code"`
	Cursor *myPaymentOrdersCursor `json:"cursor"`
}

type myPaymentOrdersCursor struct {
	CreatedAt time.Time `json:"createdAt"`
	OrderNo   string    `json:"orderNo"`
}

type orderSnapshot struct {
	Items []model.PaymentOrderItem
	Total int64
}

const maxInt64 int64 = 1<<63 - 1
const myPaymentOrdersPageSize = 10

func normalizeOrderFulfillmentType(value model.OrderFulfillmentType) (model.OrderFulfillmentType, bool) {
	if value == "" {
		return model.OrderFulfillmentDineIn, true
	}
	if value == model.OrderFulfillmentDineIn || value == model.OrderFulfillmentDelivery {
		return value, true
	}
	return "", false
}

func RegisterPaymentRoutes(app *iris.Application, db *gorm.DB, gateway *payment.Gateway) {
	gateways := make(map[string]*payment.Gateway)
	if gateway != nil {
		gateways[gateway.MerchantID()] = gateway
	}
	registerPaymentRoutes(app, db, gateways)
}

func RegisterPaymentRoutesForMerchants(app *iris.Application, db *gorm.DB, gateways map[string]*payment.Gateway) {
	registerPaymentRoutes(app, db, gateways)
}

func registerPaymentRoutes(app *iris.Application, db *gorm.DB, gateways map[string]*payment.Gateway) {
	registerAddressRoutes(app, db, gateways)
	readWechatLoginRequest := func(ctx iris.Context) (wechatLoginRequest, bool) {
		var request wechatLoginRequest
		if err := ctx.ReadJSON(&request); err != nil || request.Code == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "WeChat login code is required"})
			return wechatLoginRequest{}, false
		}
		return request, true
	}
	handleWechatLogin := func(ctx iris.Context, gateway *payment.Gateway, request wechatLoginRequest) {
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
		ctx.JSON(iris.Map{"loggedIn": true})
	}
	app.Post("/api/merchants/{merchantId:string}/auth/wechat/login", func(ctx iris.Context) {
		gateway := gateways[ctx.Params().Get("merchantId")]
		if gateway == nil {
			ctx.NotFound()
			return
		}
		request, ok := readWechatLoginRequest(ctx)
		if !ok {
			return
		}
		handleWechatLogin(ctx, gateway, request)
	})
	app.Post("/api/auth/wechat/login", func(ctx iris.Context) {
		if len(gateways) == 0 {
			ctx.StatusCode(http.StatusServiceUnavailable)
			ctx.JSON(iris.Map{"error": "WeChat login is not configured"})
			return
		}
		request, ok := readWechatLoginRequest(ctx)
		if !ok {
			return
		}
		gateway := gateways[request.MerchantID]
		if request.MerchantID == "" && len(gateways) == 1 {
			for _, configuredGateway := range gateways {
				gateway = configuredGateway
			}
		}
		if gateway == nil && request.MerchantID != "" {
			ctx.NotFound()
			return
		}
		if gateway == nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "merchant ID is required when multiple merchants are configured"})
			return
		}
		handleWechatLogin(ctx, gateway, request)
	})

	app.Post("/api/merchants/{merchantId:string}/orders/mine", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return
		}
		var request myPaymentOrdersRequest
		if err := ctx.ReadJSON(&request); err != nil || request.Code == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "WeChat login code is required"})
			return
		}
		if request.Cursor != nil && (request.Cursor.CreatedAt.IsZero() || request.Cursor.OrderNo == "") {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid order cursor"})
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

		query := db.Select("order_no", "store_id", "store_name", "fulfillment_type", "status", "total_amount", "discount_amount", "currency", "items", "created_at", "paid_at").
			Where("merchant_id = ? AND payer_open_id = ?", merchantID, openID)
		if request.Cursor != nil {
			query = query.Where("((created_at < ?) OR (created_at = ? AND order_no < ?))",
				request.Cursor.CreatedAt, request.Cursor.CreatedAt, request.Cursor.OrderNo)
		}
		var orders []model.PaymentOrder
		if err := query.Order("created_at DESC, order_no DESC").Limit(myPaymentOrdersPageSize).Find(&orders).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load user orders"})
			return
		}
		var stores []model.Store
		if err := db.Select("store_id", "name").Where("merchant_id = ?", merchantID).Find(&stores).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load store names"})
			return
		}
		storeNames := make(map[string]string, len(stores))
		for _, store := range stores {
			storeNames[store.StoreID] = store.Name
		}

		result := make([]iris.Map, 0, len(orders))
		for _, order := range orders {
			storeName := order.StoreName
			if storeName == "" {
				storeName = storeNames[order.StoreID]
			}
			result = append(result, iris.Map{
				"orderNo":         order.OrderNo,
				"storeId":         order.StoreID,
				"storeName":       storeName,
				"fulfillmentType": order.FulfillmentType,
				"status":          order.Status,
				"totalAmount":     order.TotalAmount,
				"discountAmount":  order.DiscountAmount,
				"currency":        order.Currency,
				"items":           order.Items,
				"createdAt":       order.CreatedAt,
				"paidAt":          order.PaidAt,
			})
		}
		var nextCursor *myPaymentOrdersCursor
		if len(orders) > 0 {
			lastOrder := orders[len(orders)-1]
			nextCursor = &myPaymentOrdersCursor{CreatedAt: lastOrder.CreatedAt, OrderNo: lastOrder.OrderNo}
		}
		ctx.JSON(iris.Map{"orders": result, "hasMore": len(orders) == myPaymentOrdersPageSize, "nextCursor": nextCursor})
	})

	app.Post("/api/merchants/{merchantId:string}/carts/{cartId:string}/orders", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return
		}
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		if store.IsTest {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(iris.Map{"error": "test stores cannot accept payments"})
			return
		}
		cartID := ctx.Params().Get("cartId")
		var request createPaymentOrderRequest
		if err := ctx.ReadJSON(&request); err != nil || request.Code == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "WeChat login code is required"})
			return
		}
		fulfillmentType, validFulfillmentType := normalizeOrderFulfillmentType(request.FulfillmentType)
		if !validFulfillmentType {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "fulfillmentType must be delivery or dine_in"})
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
		if fulfillmentType == model.OrderFulfillmentDelivery {
			var addressCount int64
			if err := db.Model(&model.UserAddress{}).
				Where("merchant_id = ? AND app_id = ? AND open_id = ?", merchantID, gateway.AppID(), openID).
				Count(&addressCount).Error; err != nil {
				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(iris.Map{"error": "failed to check delivery address"})
				return
			}
			if addressCount == 0 {
				ctx.StatusCode(http.StatusBadRequest)
				ctx.JSON(iris.Map{"error": "delivery requires a saved address"})
				return
			}
		}

		var order model.PaymentOrder
		err = db.Transaction(func(tx *gorm.DB) error {
			var cartItems []model.CartItem
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("merchant_id = ? AND store_id = ? AND cart_id = ?", merchantID, store.StoreID, cartID).
				Order("id ASC").Find(&cartItems).Error; err != nil {
				return err
			}
			snapshot, err := snapshotCart(cartItems)
			if err != nil {
				return err
			}
			discountAmount := int64(0)
			var benefit model.MemberBenefit
			if request.BenefitID != nil {
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("id = ? AND merchant_id = ? AND app_id = ? AND open_id = ? AND kind = ? AND status = ?", *request.BenefitID, merchantID, gateway.AppID(), openID, "coupon", "available").
					First(&benefit).Error; err != nil {
					return errInvalidCoupon
				}
				if (benefit.ExpiresAt != nil && !benefit.ExpiresAt.After(time.Now())) || snapshot.Total < benefit.MinimumAmount || benefit.DiscountAmount < 1 || benefit.DiscountAmount >= snapshot.Total {
					return errInvalidCoupon
				}
				discountAmount = benefit.DiscountAmount
			}
			orderNo, err := newOrderNo()
			if err != nil {
				return err
			}
			expiresAt := time.Now().Add(15 * time.Minute)
			order = model.PaymentOrder{
				MerchantID:      merchantID,
				StoreID:         store.StoreID,
				StoreName:       store.Name,
				FulfillmentType: fulfillmentType,
				CartID:          cartID,
				OrderNo:         orderNo,
				Status:          model.PaymentOrderStatusPending,
				TotalAmount:     snapshot.Total - discountAmount,
				DiscountAmount:  discountAmount,
				MemberBenefitID: request.BenefitID,
				Currency:        "CNY",
				Items:           snapshot.Items,
				PayerOpenID:     openID,
				ExpiresAt:       &expiresAt,
			}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
			if request.BenefitID != nil {
				result := tx.Model(&benefit).Where("status = ?", "available").Updates(map[string]any{"status": "reserved", "order_no": orderNo, "reserved_until": expiresAt})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return errInvalidCoupon
				}
			}
			return nil
		})
		if err != nil {
			if errors.Is(err, errEmptyCart) {
				ctx.StatusCode(http.StatusBadRequest)
				ctx.JSON(iris.Map{"error": "cart is empty"})
				return
			}
			if errors.Is(err, errInvalidCartTotal) {
				ctx.StatusCode(http.StatusBadRequest)
				ctx.JSON(iris.Map{"error": "cart contains an invalid amount"})
				return
			}
			if errors.Is(err, errInvalidCoupon) {
				ctx.StatusCode(http.StatusBadRequest)
				ctx.JSON(iris.Map{"error": "coupon is expired, unavailable, or does not meet the order minimum"})
				return
			}
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to create payment order"})
			return
		}

		paymentParams, prepayID, err := gateway.CreateJSAPI(ctx.Request().Context(), order.OrderNo, "知甜订单", order.TotalAmount, openID)
		if err != nil {
			log.Printf("WeChat JSAPI prepay failed for order %s: %v", order.OrderNo, err)
			_ = db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Model(&order).Update("status", model.PaymentOrderStatusClosed).Error; err != nil {
					return err
				}
				if order.MemberBenefitID != nil {
					return tx.Model(&model.MemberBenefit{}).Where("id = ? AND status = ? AND order_no = ?", *order.MemberBenefitID, "reserved", order.OrderNo).
						Updates(map[string]any{"status": "available", "order_no": "", "reserved_until": nil}).Error
				}
				return nil
			})
			ctx.StatusCode(http.StatusBadGateway)
			ctx.JSON(iris.Map{"error": "failed to create WeChat prepayment"})
			return
		}
		if err := db.Model(&order).Update("WeChatPrepayID", prepayID).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save prepayment reference"})
			return
		}
		ctx.JSON(iris.Map{"orderNo": order.OrderNo, "payment": paymentParams})
	})

	app.Get("/api/merchants/{merchantId:string}/carts/{cartId:string}/orders/{orderNo:string}", func(ctx iris.Context) {
		merchantID := ctx.Params().Get("merchantId")
		store, ok := requireActiveStore(ctx, db, merchantID)
		if !ok {
			return
		}
		var order model.PaymentOrder
		err := db.Select("merchant_id", "store_id", "store_name", "fulfillment_type", "cart_id", "order_no", "status", "total_amount", "discount_amount", "currency", "items", "paid_at").
			Where("merchant_id = ? AND store_id = ? AND cart_id = ? AND order_no = ?", merchantID, store.StoreID, ctx.Params().Get("cartId"), ctx.Params().Get("orderNo")).
			First(&order).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.NotFound()
			return
		}
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load payment order"})
			return
		}
		ctx.JSON(iris.Map{
			"orderNo":         order.OrderNo,
			"storeId":         order.StoreID,
			"storeName":       order.StoreName,
			"fulfillmentType": order.FulfillmentType,
			"status":          order.Status,
			"totalAmount":     order.TotalAmount,
			"discountAmount":  order.DiscountAmount,
			"currency":        order.Currency,
			"items":           order.Items,
			"paidAt":          order.PaidAt,
		})
	})

	app.Post("/api/payment/wechat/notify/{merchantId:string}", func(ctx iris.Context) {
		gateway := gateways[ctx.Params().Get("merchantId")]
		if gateway == nil {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "not found"})
			return
		}
		transaction, err := gateway.ParseNotification(ctx.Request().Context(), ctx.Request())
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "invalid notification"})
			return
		}
		if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
			ctx.JSON(iris.Map{"code": "SUCCESS", "message": "成功"})
			return
		}
		if transaction.OutTradeNo == nil || transaction.TransactionId == nil || transaction.Mchid == nil || *transaction.Mchid != gateway.MchID() || transaction.Appid == nil || *transaction.Appid != gateway.AppID() || transaction.Amount == nil || transaction.Amount.Total == nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "invalid transaction"})
			return
		}
		var order model.PaymentOrder
		if err := db.Where("merchant_id = ? AND order_no = ?", gateway.MerchantID(), *transaction.OutTradeNo).First(&order).Error; err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "order not found"})
			return
		}
		if order.TotalAmount != *transaction.Amount.Total || order.Currency != "CNY" || transaction.Payer == nil || transaction.Payer.Openid == nil || order.PayerOpenID != *transaction.Payer.Openid {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "transaction does not match order"})
			return
		}
		if order.Status == model.PaymentOrderStatusPaid {
			if order.WeChatTransactionID != nil && *order.WeChatTransactionID == *transaction.TransactionId {
				ctx.JSON(iris.Map{"code": "SUCCESS", "message": "成功"})
				return
			}
			ctx.StatusCode(http.StatusConflict)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "order already paid"})
			return
		}
		if order.Status != model.PaymentOrderStatusPending {
			ctx.StatusCode(http.StatusConflict)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "order is not payable"})
			return
		}

		paidAt := time.Now()
		err = db.Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&model.PaymentOrder{}).
				Where("merchant_id = ? AND order_no = ? AND status = ?", order.MerchantID, order.OrderNo, model.PaymentOrderStatusPending).
				Updates(&model.PaymentOrder{
					Status:              model.PaymentOrderStatusPaid,
					WeChatTransactionID: transaction.TransactionId,
					PaidAt:              &paidAt,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errors.New("payment order status changed")
			}
			if err := removePurchasedCartItems(tx, order); err != nil {
				return err
			}
			if err := recordMemberPurchase(tx, order, gateway.AppID()); err != nil {
				return err
			}
			if order.MemberBenefitID != nil {
				result := tx.Model(&model.MemberBenefit{}).Where("id = ? AND status = ? AND order_no = ?", *order.MemberBenefitID, "reserved", order.OrderNo).
					Updates(map[string]any{"status": "used", "reserved_until": nil})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return errors.New("coupon reservation changed")
				}
			}
			return nil
		})
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"code": "FAIL", "message": "failed to update order"})
			return
		}
		ctx.JSON(iris.Map{"code": "SUCCESS", "message": "成功"})
	})
}

func persistWeChatUser(db *gorm.DB, appID, openID string) error {
	user := model.WeChatUser{AppID: appID, OpenID: openID}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app_id"}, {Name: "openid"}},
		DoNothing: true,
	}).Create(&user).Error
}

var (
	errEmptyCart        = errors.New("cart is empty")
	errInvalidCartTotal = errors.New("cart contains an invalid amount")
	errInvalidCoupon    = errors.New("coupon is unavailable or does not meet the order minimum")
)

func snapshotCart(cartItems []model.CartItem) (orderSnapshot, error) {
	if len(cartItems) == 0 {
		return orderSnapshot{}, errEmptyCart
	}
	snapshot := orderSnapshot{Items: make([]model.PaymentOrderItem, 0, len(cartItems))}
	for _, item := range cartItems {
		unitAmount, err := priceToFen(item.UnitPrice)
		if err != nil || item.Quantity < 1 || item.Quantity > 99 || unitAmount > maxInt64/int64(item.Quantity) {
			return orderSnapshot{}, errInvalidCartTotal
		}
		lineAmount := unitAmount * int64(item.Quantity)
		if snapshot.Total > maxInt64-lineAmount {
			return orderSnapshot{}, errInvalidCartTotal
		}
		snapshot.Total += lineAmount
		snapshot.Items = append(snapshot.Items, model.PaymentOrderItem{
			CartItemID: item.ID,
			ProductID:  item.ProductID,
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitAmount: unitAmount,
			LineAmount: lineAmount,
			Options:    item.Options,
		})
	}
	if snapshot.Total < 1 {
		return orderSnapshot{}, errInvalidCartTotal
	}
	return snapshot, nil
}

func priceToFen(price float64) (int64, error) {
	if math.IsNaN(price) || math.IsInf(price, 0) || price < 0 || price > float64(maxInt64)/100 {
		return 0, errInvalidCartTotal
	}
	amount := math.Round(price * 100)
	if amount >= float64(maxInt64) {
		return 0, errInvalidCartTotal
	}
	return int64(amount), nil
}

func newOrderNo() (string, error) {
	var random [9]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return "MD" + time.Now().UTC().Format("060102150405") + hex.EncodeToString(random[:]), nil
}

func removePurchasedCartItems(tx *gorm.DB, order model.PaymentOrder) error {
	for _, purchased := range order.Items {
		var current model.CartItem
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("merchant_id = ? AND store_id = ? AND cart_id = ? AND id = ?", order.MerchantID, order.StoreID, order.CartID, purchased.CartItemID).
			First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if current.Quantity > purchased.Quantity {
			if err := tx.Model(&current).Update("quantity", current.Quantity-purchased.Quantity).Error; err != nil {
				return err
			}
			continue
		}
		if err := tx.Delete(&current).Error; err != nil {
			return err
		}
	}
	return nil
}
