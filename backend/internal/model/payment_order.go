package model

import "time"

// PaymentOrder 表示商户的一笔支付订单及创建时的商品、金额快照，金额以人民币分为单位。
type PaymentOrder struct {
	ID                  uint                 `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	MerchantID          string               `json:"merchantId" gorm:"size:64;not null;uniqueIndex:idx_payment_orders_merchant_order_no,priority:1;index:idx_payment_orders_merchant_status,priority:1;index:idx_payment_orders_merchant_store,priority:1"`
	StoreID             string               `json:"storeId" gorm:"size:64;not null;default:default;index:idx_payment_orders_merchant_store,priority:2"`
	StoreName           string               `json:"storeName" gorm:"size:128;not null;default:''"`
	FulfillmentType     OrderFulfillmentType `json:"fulfillmentType" gorm:"size:16;not null;default:dine_in"`
	CartID              string               `json:"-" gorm:"size:64;not null;index"`
	OrderNo             string               `json:"orderNo" gorm:"size:64;not null;uniqueIndex:idx_payment_orders_merchant_order_no,priority:2"`
	Status              PaymentOrderStatus   `json:"status" gorm:"size:16;not null;default:pending;index:idx_payment_orders_merchant_status,priority:2"`
	TotalAmount         int64                `json:"totalAmount" gorm:"not null"`
	DiscountAmount      int64                `json:"discountAmount" gorm:"not null;default:0"`
	MemberBenefitID     *uint                `json:"memberBenefitId,omitempty" gorm:"index"`
	Currency            string               `json:"currency" gorm:"size:3;not null;default:CNY"`
	Items               []PaymentOrderItem   `json:"items" gorm:"serializer:json;type:text;not null"`
	WeChatTransactionID *string              `json:"wechatTransactionId,omitempty" gorm:"size:64;uniqueIndex"`
	WeChatPrepayID      string               `json:"-" gorm:"size:128"`
	PayerOpenID         string               `json:"-" gorm:"size:128;index"`
	ExpiresAt           *time.Time           `json:"expiresAt,omitempty"`
	PaidAt              *time.Time           `json:"paidAt,omitempty"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

type OrderFulfillmentType string

const (
	OrderFulfillmentDelivery OrderFulfillmentType = "delivery"
	OrderFulfillmentDineIn   OrderFulfillmentType = "dine_in"
)

// PaymentOrderStatus 表示支付订单的业务状态。
type PaymentOrderStatus string

const (
	PaymentOrderStatusPending  PaymentOrderStatus = "pending"
	PaymentOrderStatusPaid     PaymentOrderStatus = "paid"
	PaymentOrderStatusClosed   PaymentOrderStatus = "closed"
	PaymentOrderStatusRefunded PaymentOrderStatus = "refunded"
)

// PaymentOrderItem 表示下单时保存的商品信息和分币金额快照。
type PaymentOrderItem struct {
	CartItemID uint   `json:"cartItemId"`
	ProductID  string `json:"productId"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitAmount int64  `json:"unitAmount"`
	LineAmount int64  `json:"lineAmount"`
	Options    string `json:"options"`
}

// TableName 指定支付订单在数据库中的表名。
func (PaymentOrder) TableName() string {
	return "payment_orders"
}
