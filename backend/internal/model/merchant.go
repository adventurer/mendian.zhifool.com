package model

import "time"

// Merchant 表示一个商户及其微信支付配置。
type Merchant struct {
	MerchantID                 string    `json:"merchantId" gorm:"primaryKey;size:64"`
	Name                       string    `json:"name" gorm:"size:128;not null"`
	WeChatPayEnabled           bool      `json:"wechatPayEnabled" gorm:"not null;default:false"`
	WeChatAppID                string    `json:"wechatAppId" gorm:"size:128"`
	WeChatMchID                string    `json:"wechatMchId" gorm:"size:32;index"`
	WeChatCertificateSerialNo  string    `json:"wechatCertificateSerialNo" gorm:"size:64"`
	WeChatAPIv3KeyCiphertext   string    `json:"-" gorm:"type:text"`
	WeChatPrivateKeyCiphertext string    `json:"-" gorm:"type:text"`
	WeChatNotifyURL            string    `json:"wechatNotifyUrl" gorm:"size:512"`
	CreatedAt                  time.Time `json:"createdAt"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// TableName 指定商户数据在数据库中的表名。
func (Merchant) TableName() string {
	return "merchants"
}
