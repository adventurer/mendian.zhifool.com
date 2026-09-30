package model

import "time"

type MemberProfile struct {
	MerchantID string    `json:"-" gorm:"primaryKey;size:64"`
	AppID      string    `json:"-" gorm:"primaryKey;size:64"`
	OpenID     string    `json:"-" gorm:"primaryKey;size:128"`
	MemberNo   string    `json:"memberNo" gorm:"size:32;not null;uniqueIndex"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (MemberProfile) TableName() string { return "member_profiles" }

type MemberPointsLedger struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	MerchantID     string    `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_points_source,priority:1;index:idx_member_points_owner,priority:1"`
	AppID          string    `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_points_source,priority:2;index:idx_member_points_owner,priority:2"`
	OpenID         string    `json:"-" gorm:"size:128;not null;uniqueIndex:idx_member_points_source,priority:3;index:idx_member_points_owner,priority:3"`
	SourceType     string    `json:"sourceType" gorm:"size:24;not null;uniqueIndex:idx_member_points_source,priority:4"`
	SourceID       string    `json:"sourceId" gorm:"size:64;not null;uniqueIndex:idx_member_points_source,priority:5"`
	RelatedOrderNo string    `json:"relatedOrderNo,omitempty" gorm:"size:64;index"`
	Points         int64     `json:"points" gorm:"not null"`
	Description    string    `json:"description" gorm:"size:255;not null"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (MemberPointsLedger) TableName() string { return "member_points_ledger" }

type MemberMessage struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	MerchantID string    `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_message_source,priority:1;index:idx_member_messages_owner,priority:1"`
	AppID      string    `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_message_source,priority:2;index:idx_member_messages_owner,priority:2"`
	OpenID     string    `json:"-" gorm:"size:128;not null;uniqueIndex:idx_member_message_source,priority:3;index:idx_member_messages_owner,priority:3"`
	SourceType string    `json:"sourceType" gorm:"size:24;not null;uniqueIndex:idx_member_message_source,priority:4"`
	SourceID   string    `json:"sourceId" gorm:"size:64;not null;uniqueIndex:idx_member_message_source,priority:5"`
	Title      string    `json:"title" gorm:"size:128;not null"`
	Content    string    `json:"content" gorm:"size:512;not null"`
	IsRead     bool      `json:"isRead" gorm:"not null;default:false;index"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (MemberMessage) TableName() string { return "member_messages" }

type InvoiceRequest struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	MerchantID string    `json:"-" gorm:"size:64;not null;uniqueIndex:idx_invoice_order,priority:1;index:idx_invoice_owner,priority:1"`
	AppID      string    `json:"-" gorm:"size:64;not null;index:idx_invoice_owner,priority:2"`
	OpenID     string    `json:"-" gorm:"size:128;not null;uniqueIndex:idx_invoice_order,priority:3;index:idx_invoice_owner,priority:3"`
	OrderNo    string    `json:"orderNo" gorm:"size:64;not null;uniqueIndex:idx_invoice_order,priority:2"`
	Title      string    `json:"title" gorm:"size:128;not null"`
	TaxNumber  string    `json:"taxNumber" gorm:"size:32"`
	Email      string    `json:"email" gorm:"size:255;not null"`
	Status     string    `json:"status" gorm:"size:16;not null;default:pending"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (InvoiceRequest) TableName() string { return "invoice_requests" }

type MemberBenefitCode struct {
	MerchantID     string     `json:"-" gorm:"primaryKey;size:64"`
	Code           string     `json:"code" gorm:"primaryKey;size:64"`
	StoreID        string     `json:"storeId,omitempty" gorm:"size:64;not null;default:'';index"`
	Kind           string     `json:"kind" gorm:"size:16;not null"`
	Title          string     `json:"title" gorm:"size:128;not null"`
	Description    string     `json:"description" gorm:"size:255;not null;default:''"`
	DiscountAmount int64      `json:"discountAmount" gorm:"not null;default:0"`
	MinimumAmount  int64      `json:"minimumAmount" gorm:"not null;default:0"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	IsActive       bool       `json:"isActive" gorm:"not null;default:true"`
	MaxClaims      int64      `json:"maxClaims" gorm:"not null;default:0"`
	ClaimsCount    int64      `json:"claimsCount" gorm:"not null;default:0"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func (MemberBenefitCode) TableName() string { return "member_benefit_codes" }

type MemberBenefit struct {
	ID             uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	ClaimNo        string     `json:"claimNo" gorm:"size:32;uniqueIndex"`
	MerchantID     string     `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_benefit_source,priority:1;index:idx_member_benefit_owner,priority:1"`
	StoreID        string     `json:"storeId,omitempty" gorm:"size:64;not null;default:'';index"`
	AppID          string     `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_benefit_source,priority:2;index:idx_member_benefit_owner,priority:2"`
	OpenID         string     `json:"-" gorm:"size:128;not null;uniqueIndex:idx_member_benefit_source,priority:3;index:idx_member_benefit_owner,priority:3"`
	SourceCode     string     `json:"sourceCode" gorm:"size:64;not null;uniqueIndex:idx_member_benefit_source,priority:4"`
	Kind           string     `json:"kind" gorm:"size:16;not null;index"`
	Title          string     `json:"title" gorm:"size:128;not null"`
	Description    string     `json:"description" gorm:"size:255;not null;default:''"`
	DiscountAmount int64      `json:"discountAmount" gorm:"not null;default:0"`
	MinimumAmount  int64      `json:"minimumAmount" gorm:"not null;default:0"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	Status         string     `json:"status" gorm:"size:16;not null;default:available;index"`
	OrderNo        string     `json:"orderNo,omitempty" gorm:"size:64"`
	ReservedUntil  *time.Time `json:"-"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func (MemberBenefit) TableName() string { return "member_benefits" }

type MemberRefund struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	MerchantID string    `json:"-" gorm:"size:64;not null;uniqueIndex:idx_member_refund_no,priority:1;index:idx_member_refunds_order,priority:1"`
	RefundNo   string    `json:"refundNo" gorm:"size:64;not null;uniqueIndex:idx_member_refund_no,priority:2"`
	OrderNo    string    `json:"orderNo" gorm:"size:64;not null;index:idx_member_refunds_order,priority:2"`
	Amount     int64     `json:"amount" gorm:"not null"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (MemberRefund) TableName() string { return "member_refunds" }
