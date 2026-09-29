package api

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mendian-backend/internal/model"
	"mendian-backend/internal/payment"
)

type memberRequest struct {
	Code string `json:"code"`
}
type invoiceCreateRequest struct {
	Code      string `json:"code"`
	OrderNo   string `json:"orderNo"`
	Title     string `json:"title"`
	TaxNumber string `json:"taxNumber"`
	Email     string `json:"email"`
}

type benefitRequest struct {
	Code       string `json:"code"`
	RedeemCode string `json:"redeemCode"`
}

type benefitCodeRequest struct {
	Kind           string     `json:"kind"`
	Code           string     `json:"code"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	DiscountAmount int64      `json:"discountAmount"`
	MinimumAmount  int64      `json:"minimumAmount"`
	MaxClaims      int64      `json:"maxClaims"`
	ExpiresAt      *time.Time `json:"expiresAt"`
}

type invoiceStatusRequest struct {
	Status string `json:"status"`
}

type confirmedRefundRequest struct {
	RefundNo  string `json:"refundNo"`
	AmountFen int64  `json:"amountFen"`
}

func RegisterMemberRoutes(app *iris.Application, db *gorm.DB, gateways map[string]*payment.Gateway) {
	resolve := func(ctx iris.Context, code string) (*payment.Gateway, string, bool) {
		merchantID := ctx.Params().Get("merchantId")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return nil, "", false
		}
		if strings.TrimSpace(code) == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "WeChat login code is required"})
			return nil, "", false
		}
		openID, err := gateway.ResolveOpenID(ctx.Request().Context(), code)
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "WeChat login failed; please retry"})
			return nil, "", false
		}
		if err := persistWeChatUser(db, gateway.AppID(), openID); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to save WeChat user"})
			return nil, "", false
		}
		return gateway, openID, true
	}
	app.Post("/api/merchants/{merchantId:string}/members/summary", func(ctx iris.Context) {
		var req memberRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		profile, err := getOrCreateMemberProfile(db, ctx.Params().Get("merchantId"), gateway.AppID(), openID)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "failed to load member profile"})
			return
		}
		owner := map[string]any{"merchant_id": profile.MerchantID, "app_id": profile.AppID, "open_id": profile.OpenID}
		var points, unread, coupons, gifts int64
		if err := db.Model(&model.MemberPointsLedger{}).Where(owner).Select("COALESCE(SUM(points), 0)").Scan(&points).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to load member points"})
			return
		}
		if err := db.Model(&model.MemberMessage{}).Where(owner).Where("is_read = ?", false).Count(&unread).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to load unread messages"})
			return
		}
		benefitQuery := func() *gorm.DB {
			return db.Model(&model.MemberBenefit{}).Where(owner).Where("status = ? AND (expires_at IS NULL OR expires_at > ?)", "available", time.Now())
		}
		if err := benefitQuery().Where("kind = ?", "coupon").Count(&coupons).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to count coupons"})
			return
		}
		if err := benefitQuery().Where("kind = ?", "gift").Count(&gifts).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to count gifts"})
			return
		}
		ctx.JSON(iris.Map{"memberNo": profile.MemberNo, "points": points, "unreadMessages": unread, "couponCount": coupons, "giftCount": gifts})
	})
	for _, kind := range []string{"coupon", "gift"} {
		benefitKind := kind
		app.Post("/api/merchants/{merchantId:string}/benefits/"+benefitKind+"/list", func(ctx iris.Context) {
			var req memberRequest
			if err := ctx.ReadJSON(&req); err != nil {
				ctx.StatusCode(400)
				ctx.JSON(iris.Map{"error": "invalid request body"})
				return
			}
			gateway, openID, ok := resolve(ctx, req.Code)
			if !ok {
				return
			}
			owner := map[string]any{"merchant_id": ctx.Params().Get("merchantId"), "app_id": gateway.AppID(), "open_id": openID, "kind": benefitKind}
			if err := db.Model(&model.MemberBenefit{}).Where(owner).Where("status = ? AND reserved_until < ?", "reserved", time.Now()).Updates(map[string]any{"status": "available", "order_no": "", "reserved_until": nil}).Error; err != nil {
				ctx.StatusCode(500)
				ctx.JSON(iris.Map{"error": "failed to release expired benefit reservations"})
				return
			}
			var benefits []model.MemberBenefit
			if err := db.Where(owner).Where("status IN ? AND (status = ? OR expires_at IS NULL OR expires_at > ?)", []string{"available", "used"}, "used", time.Now()).Order("created_at DESC, id DESC").Find(&benefits).Error; err != nil {
				ctx.StatusCode(500)
				ctx.JSON(iris.Map{"error": "failed to load benefits"})
				return
			}
			ctx.JSON(iris.Map{"benefits": benefits})
		})
		app.Post("/api/merchants/{merchantId:string}/benefits/"+benefitKind+"/claim", func(ctx iris.Context) {
			var req benefitRequest
			if err := ctx.ReadJSON(&req); err != nil || strings.TrimSpace(req.RedeemCode) == "" {
				ctx.StatusCode(400)
				ctx.JSON(iris.Map{"error": "redemption code is required"})
				return
			}
			gateway, openID, ok := resolve(ctx, req.Code)
			if !ok {
				return
			}
			merchantID := ctx.Params().Get("merchantId")
			var claimed model.MemberBenefit
			err := db.Transaction(func(tx *gorm.DB) error {
				var code model.MemberBenefitCode
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchant_id = ? AND code = ? AND kind = ? AND is_active = ?", merchantID, strings.TrimSpace(req.RedeemCode), benefitKind, true).First(&code).Error; err != nil {
					return err
				}
				if code.ExpiresAt != nil && !code.ExpiresAt.After(time.Now()) {
					return gorm.ErrRecordNotFound
				}
				if code.MaxClaims > 0 && code.ClaimsCount >= code.MaxClaims {
					return errors.New("benefit code exhausted")
				}
				var claimRandom [6]byte
				if _, err := rand.Read(claimRandom[:]); err != nil {
					return err
				}
				claimed = model.MemberBenefit{ClaimNo: "ZT" + strings.ToUpper(hex.EncodeToString(claimRandom[:])), MerchantID: merchantID, AppID: gateway.AppID(), OpenID: openID, SourceCode: code.Code, Kind: code.Kind, Title: code.Title, Description: code.Description, DiscountAmount: code.DiscountAmount, MinimumAmount: code.MinimumAmount, ExpiresAt: code.ExpiresAt, Status: "available"}
				if err := tx.Create(&claimed).Error; err != nil {
					return err
				}
				return tx.Model(&code).UpdateColumn("claims_count", gorm.Expr("claims_count + 1")).Error
			})
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.StatusCode(404)
				ctx.JSON(iris.Map{"error": "redemption code is invalid or expired"})
				return
			}
			if err != nil {
				if strings.Contains(err.Error(), "duplicate") {
					ctx.StatusCode(409)
					ctx.JSON(iris.Map{"error": "this code has already been claimed"})
					return
				}
				if strings.Contains(err.Error(), "exhausted") {
					ctx.StatusCode(409)
					ctx.JSON(iris.Map{"error": "redemption code has reached its limit"})
					return
				}
				ctx.StatusCode(500)
				ctx.JSON(iris.Map{"error": "failed to claim benefit"})
				return
			}
			ctx.StatusCode(http.StatusCreated)
			ctx.JSON(iris.Map{"benefit": claimed})
		})
	}
	app.Post("/api/admin/merchants/{merchantId:string}/benefit-codes", func(ctx iris.Context) {
		token := os.Getenv("MEMBER_BENEFITS_ADMIN_TOKEN")
		if token == "" {
			ctx.StatusCode(http.StatusServiceUnavailable)
			ctx.JSON(iris.Map{"error": "benefit code management is not configured"})
			return
		}
		if !hmac.Equal([]byte(ctx.GetHeader("X-Operator-Token")), []byte(token)) {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(iris.Map{"error": "unauthorized"})
			return
		}
		var req benefitCodeRequest
		if err := ctx.ReadJSON(&req); err != nil || (req.Kind != "coupon" && req.Kind != "gift") || strings.TrimSpace(req.Code) == "" || len(req.Code) > 64 || strings.TrimSpace(req.Title) == "" || len([]rune(req.Title)) > 128 || req.DiscountAmount < 0 || req.MinimumAmount < 0 || req.MaxClaims < 0 {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid benefit code"})
			return
		}
		code := model.MemberBenefitCode{MerchantID: ctx.Params().Get("merchantId"), Code: strings.TrimSpace(req.Code), Kind: req.Kind, Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description), DiscountAmount: req.DiscountAmount, MinimumAmount: req.MinimumAmount, ExpiresAt: req.ExpiresAt, MaxClaims: req.MaxClaims, IsActive: true}
		if code.Kind == "gift" && code.DiscountAmount != 0 {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "gift benefits cannot discount an order"})
			return
		}
		if code.Kind == "coupon" && code.DiscountAmount < 1 {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "coupon discountAmount must be positive"})
			return
		}
		if err := db.Create(&code).Error; err != nil {
			ctx.StatusCode(409)
			ctx.JSON(iris.Map{"error": "benefit code already exists"})
			return
		}
		ctx.StatusCode(http.StatusCreated)
		ctx.JSON(iris.Map{"code": code.Code, "kind": code.Kind, "title": code.Title})
	})
	app.Post("/api/admin/merchants/{merchantId:string}/gifts/{claimNo:string}/redeem", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		merchantID, claimNo := ctx.Params().Get("merchantId"), ctx.Params().Get("claimNo")
		var benefit model.MemberBenefit
		if err := db.Where("merchant_id = ? AND claim_no = ? AND kind = ?", merchantID, claimNo, "gift").First(&benefit).Error; err != nil {
			ctx.NotFound()
			return
		}
		if benefit.Status != "available" {
			ctx.StatusCode(http.StatusConflict)
			ctx.JSON(iris.Map{"error": "gift has already been redeemed or is unavailable"})
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&benefit).Where("status = ?", "available").Update("status", "used")
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errors.New("gift status changed")
			}
			message := model.MemberMessage{MerchantID: benefit.MerchantID, AppID: benefit.AppID, OpenID: benefit.OpenID, SourceType: "gift", SourceID: benefit.ClaimNo, Title: "礼品已核销", Content: "您的礼品「" + benefit.Title + "」已由门店核销。"}
			return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&message).Error
		}); err != nil {
			ctx.StatusCode(409)
			ctx.JSON(iris.Map{"error": "gift could not be redeemed"})
			return
		}
		ctx.JSON(iris.Map{"claimNo": benefit.ClaimNo, "status": "used"})
	})
	app.Post("/api/merchants/{merchantId:string}/members/points", func(ctx iris.Context) {
		var req memberRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		owner := map[string]any{"merchant_id": ctx.Params().Get("merchantId"), "app_id": gateway.AppID(), "open_id": openID}
		var history []model.MemberPointsLedger
		if err := db.Where(owner).Order("created_at DESC, id DESC").Limit(100).Find(&history).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to load points history"})
			return
		}
		var points int64
		if err := db.Model(&model.MemberPointsLedger{}).Where(owner).Select("COALESCE(SUM(points), 0)").Scan(&points).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to load points balance"})
			return
		}
		ctx.JSON(iris.Map{"points": points, "history": history})
	})
	app.Post("/api/merchants/{merchantId:string}/messages/list", func(ctx iris.Context) {
		var req memberRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		var messages []model.MemberMessage
		if err := db.Where("merchant_id = ? AND app_id = ? AND open_id = ?", ctx.Params().Get("merchantId"), gateway.AppID(), openID).Order("created_at DESC, id DESC").Limit(100).Find(&messages).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to load messages"})
			return
		}
		ctx.JSON(iris.Map{"messages": messages})
	})
	app.Post("/api/merchants/{merchantId:string}/messages/read-all", func(ctx iris.Context) {
		var req memberRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		if err := db.Model(&model.MemberMessage{}).Where("merchant_id = ? AND app_id = ? AND open_id = ?", ctx.Params().Get("merchantId"), gateway.AppID(), openID).Update("is_read", true).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to mark messages read"})
			return
		}
		ctx.StatusCode(http.StatusNoContent)
	})
	app.Patch("/api/merchants/{merchantId:string}/messages/{messageID:string}/read", func(ctx iris.Context) {
		var req memberRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		id, err := strconv.ParseUint(ctx.Params().Get("messageID"), 10, 64)
		if err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid message ID"})
			return
		}
		result := db.Model(&model.MemberMessage{}).Where("id = ? AND merchant_id = ? AND app_id = ? AND open_id = ?", id, ctx.Params().Get("merchantId"), gateway.AppID(), openID).Update("is_read", true)
		if result.Error != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to mark message read"})
			return
		}
		if result.RowsAffected == 0 {
			ctx.NotFound()
			return
		}
		ctx.StatusCode(http.StatusNoContent)
	})
	app.Post("/api/merchants/{merchantId:string}/invoices/list", func(ctx iris.Context) {
		var req memberRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid request body"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		var invoices []model.InvoiceRequest
		if err := db.Where("merchant_id = ? AND app_id = ? AND open_id = ?", ctx.Params().Get("merchantId"), gateway.AppID(), openID).Order("created_at DESC, id DESC").Find(&invoices).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to load invoice requests"})
			return
		}
		ctx.JSON(iris.Map{"invoices": invoices})
	})
	app.Post("/api/merchants/{merchantId:string}/invoices", func(ctx iris.Context) {
		var req invoiceCreateRequest
		if err := ctx.ReadJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" || strings.TrimSpace(req.Title) == "" || len([]rune(req.Title)) > 128 || !strings.Contains(req.Email, "@") || len(req.Email) > 255 {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "valid order, invoice title and email are required"})
			return
		}
		req.TaxNumber = strings.TrimSpace(req.TaxNumber)
		if req.TaxNumber != "" && (len(req.TaxNumber) < 6 || len(req.TaxNumber) > 32) {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "tax number must be 6 to 32 characters"})
			return
		}
		gateway, openID, ok := resolve(ctx, req.Code)
		if !ok {
			return
		}
		merchantID := ctx.Params().Get("merchantId")
		var order model.PaymentOrder
		err := db.Select("merchant_id", "order_no", "status").Where("merchant_id = ? AND order_no = ? AND payer_open_id = ? AND status = ?", merchantID, req.OrderNo, openID, model.PaymentOrderStatusPaid).First(&order).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "only your paid orders can be invoiced"})
			return
		}
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to check order"})
			return
		}
		invoice := model.InvoiceRequest{MerchantID: merchantID, AppID: gateway.AppID(), OpenID: openID, OrderNo: req.OrderNo, Title: strings.TrimSpace(req.Title), TaxNumber: req.TaxNumber, Email: strings.TrimSpace(req.Email), Status: "pending"}
		if err := db.Create(&invoice).Error; err != nil {
			ctx.StatusCode(409)
			ctx.JSON(iris.Map{"error": "this order already has an invoice request"})
			return
		}
		ctx.StatusCode(http.StatusCreated)
		ctx.JSON(invoice)
	})
	app.Patch("/api/admin/merchants/{merchantId:string}/invoices/{invoiceID:string}", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		var req invoiceStatusRequest
		if err := ctx.ReadJSON(&req); err != nil || (req.Status != "processing" && req.Status != "issued" && req.Status != "rejected") {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invoice status must be processing, issued, or rejected"})
			return
		}
		id, err := strconv.ParseUint(ctx.Params().Get("invoiceID"), 10, 64)
		if err != nil {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "invalid invoice ID"})
			return
		}
		var invoice model.InvoiceRequest
		if err := db.Where("id = ? AND merchant_id = ?", id, ctx.Params().Get("merchantId")).First(&invoice).Error; err != nil {
			ctx.NotFound()
			return
		}
		if err := db.Model(&invoice).Update("status", req.Status).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to update invoice status"})
			return
		}
		title, content := "发票申请处理中", "您的发票申请正在处理中。"
		if req.Status == "issued" {
			title, content = "发票已开具", "您的发票已开具并发送至申请邮箱。"
		}
		if req.Status == "rejected" {
			title, content = "发票申请需要补充信息", "请联系门店补充发票信息。"
		}
		message := model.MemberMessage{MerchantID: invoice.MerchantID, AppID: invoice.AppID, OpenID: invoice.OpenID, SourceType: "invoice", SourceID: strconv.FormatUint(id, 10) + "_" + req.Status, Title: title, Content: content}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&message).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to notify member"})
			return
		}
		invoice.Status = req.Status
		ctx.JSON(invoice)
	})
	app.Post("/api/admin/merchants/{merchantId:string}/orders/{orderNo:string}/refunds/confirmed", func(ctx iris.Context) {
		if !authorizeMemberOperator(ctx) {
			return
		}
		var req confirmedRefundRequest
		if err := ctx.ReadJSON(&req); err != nil || strings.TrimSpace(req.RefundNo) == "" || len(req.RefundNo) > 64 || req.AmountFen < 1 {
			ctx.StatusCode(400)
			ctx.JSON(iris.Map{"error": "refund number and positive amountFen are required"})
			return
		}
		merchantID, orderNo := ctx.Params().Get("merchantId"), ctx.Params().Get("orderNo")
		gateway := gateways[merchantID]
		if gateway == nil {
			ctx.NotFound()
			return
		}
		var reversedPoints int64
		err := db.Transaction(func(tx *gorm.DB) error {
			var order model.PaymentOrder
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchant_id = ? AND order_no = ?", merchantID, orderNo).First(&order).Error; err != nil {
				return err
			}
			if order.Status != model.PaymentOrderStatusPaid && order.Status != model.PaymentOrderStatusRefunded {
				return errors.New("order is not refundable")
			}
			var refunded int64
			if err := tx.Model(&model.MemberRefund{}).Where("merchant_id = ? AND order_no = ?", merchantID, orderNo).Select("COALESCE(SUM(amount), 0)").Scan(&refunded).Error; err != nil {
				return err
			}
			if req.AmountFen > order.TotalAmount-refunded {
				return errors.New("refund amount exceeds remaining order amount")
			}
			refund := model.MemberRefund{MerchantID: merchantID, RefundNo: strings.TrimSpace(req.RefundNo), OrderNo: orderNo, Amount: req.AmountFen}
			if err := tx.Create(&refund).Error; err != nil {
				return err
			}
			refunded += req.AmountFen
			if refunded == order.TotalAmount {
				if err := tx.Model(&order).Update("status", model.PaymentOrderStatusRefunded).Error; err != nil {
					return err
				}
			}
			target := refunded / 100
			if target > order.TotalAmount/100 {
				target = order.TotalAmount / 100
			}
			var alreadyReversed int64
			if err := tx.Model(&model.MemberPointsLedger{}).Where("merchant_id = ? AND app_id = ? AND open_id = ? AND source_type = ? AND related_order_no = ?", merchantID, gateway.AppID(), order.PayerOpenID, "refund", orderNo).Select("COALESCE(SUM(-points), 0)").Scan(&alreadyReversed).Error; err != nil {
				return err
			}
			delta := target - alreadyReversed
			if delta > 0 {
				ledger := model.MemberPointsLedger{MerchantID: merchantID, AppID: gateway.AppID(), OpenID: order.PayerOpenID, SourceType: "refund", SourceID: refund.RefundNo, RelatedOrderNo: orderNo, Points: -delta, Description: "退款扣回积分"}
				if err := tx.Create(&ledger).Error; err != nil {
					return err
				}
				reversedPoints = delta
			}
			message := model.MemberMessage{MerchantID: merchantID, AppID: gateway.AppID(), OpenID: order.PayerOpenID, SourceType: "refund", SourceID: refund.RefundNo, Title: "订单退款已确认", Content: "订单 " + orderNo + " 的退款已确认，已扣回 " + strconv.FormatInt(reversedPoints, 10) + " 积分。"}
			return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&message).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.NotFound()
			return
		}
		if err != nil {
			if strings.Contains(err.Error(), "not refundable") || strings.Contains(err.Error(), "exceeds remaining") {
				ctx.StatusCode(409)
				ctx.JSON(iris.Map{"error": err.Error()})
				return
			}
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				ctx.StatusCode(409)
				ctx.JSON(iris.Map{"error": "refund number already recorded"})
				return
			}
			ctx.StatusCode(500)
			ctx.JSON(iris.Map{"error": "failed to record confirmed refund"})
			return
		}
		ctx.JSON(iris.Map{"recorded": true, "pointsReversed": reversedPoints})
	})
}

func authorizeMemberOperator(ctx iris.Context) bool {
	token := os.Getenv("MEMBER_BENEFITS_ADMIN_TOKEN")
	if token == "" {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(iris.Map{"error": "member operator API is not configured"})
		return false
	}
	if !hmac.Equal([]byte(ctx.GetHeader("X-Operator-Token")), []byte(token)) {
		ctx.StatusCode(http.StatusUnauthorized)
		ctx.JSON(iris.Map{"error": "unauthorized"})
		return false
	}
	return true
}

func getOrCreateMemberProfile(db *gorm.DB, merchantID, appID, openID string) (model.MemberProfile, error) {
	var random [5]byte
	if _, err := rand.Read(random[:]); err != nil {
		return model.MemberProfile{}, err
	}
	profile := model.MemberProfile{MerchantID: merchantID, AppID: appID, OpenID: openID, MemberNo: "ZT" + strings.ToUpper(hex.EncodeToString(random[:]))}
	err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "merchant_id"}, {Name: "app_id"}, {Name: "open_id"}}, DoNothing: true}).Create(&profile).Error
	if err != nil {
		return model.MemberProfile{}, err
	}
	err = db.Where("merchant_id = ? AND app_id = ? AND open_id = ?", merchantID, appID, openID).First(&profile).Error
	return profile, err
}

func recordMemberPurchase(tx *gorm.DB, order model.PaymentOrder, appID string) error {
	points := order.TotalAmount / 100
	if points > 0 {
		ledger := model.MemberPointsLedger{MerchantID: order.MerchantID, AppID: appID, OpenID: order.PayerOpenID, SourceType: "order_paid", SourceID: order.OrderNo, Points: points, Description: "订单消费积分"}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&ledger).Error; err != nil {
			return err
		}
	}
	message := model.MemberMessage{MerchantID: order.MerchantID, AppID: appID, OpenID: order.PayerOpenID, SourceType: "order_paid", SourceID: order.OrderNo, Title: "订单支付成功", Content: "订单 " + order.OrderNo + " 已支付成功，可在我的订单中查看详情。"}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&message).Error
}
