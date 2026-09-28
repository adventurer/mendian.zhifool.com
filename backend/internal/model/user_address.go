package model

import "time"

type UserAddress struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	MerchantID string    `json:"-" gorm:"size:64;not null;index:idx_user_addresses_owner,priority:1"`
	AppID      string    `json:"-" gorm:"size:64;not null;index:idx_user_addresses_owner,priority:2"`
	OpenID     string    `json:"-" gorm:"size:128;not null;index:idx_user_addresses_owner,priority:3"`
	Recipient  string    `json:"recipient" gorm:"size:64;not null"`
	Phone      string    `json:"phone" gorm:"size:24;not null"`
	Province   string    `json:"province" gorm:"size:64;not null;default:''"`
	City       string    `json:"city" gorm:"size:64;not null;default:''"`
	District   string    `json:"district" gorm:"size:64;not null;default:''"`
	Detail     string    `json:"detail" gorm:"size:255;not null"`
	IsDefault  bool      `json:"isDefault" gorm:"not null;default:false;index:idx_user_addresses_owner,priority:4"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (UserAddress) TableName() string {
	return "user_addresses"
}
