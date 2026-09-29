package model

import "time"

const DefaultStoreID = "default"

// Store represents a physical location belonging to a merchant.
type Store struct {
	MerchantID string    `json:"merchantId" gorm:"primaryKey;size:64"`
	StoreID    string    `json:"storeId" gorm:"primaryKey;size:64"`
	Name       string    `json:"name" gorm:"size:128;not null"`
	Phone      string    `json:"phone" gorm:"size:24;not null;default:''"`
	Address    string    `json:"address" gorm:"size:255;not null;default:''"`
	Latitude   float64   `json:"latitude" gorm:"not null;default:0"`
	Longitude  float64   `json:"longitude" gorm:"not null;default:0"`
	IsActive   bool      `json:"isActive" gorm:"not null;default:true;index"`
	IsDefault  bool      `json:"isDefault" gorm:"not null;default:false"`
	IsTest     bool      `json:"isTest" gorm:"not null;default:false"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Store) TableName() string {
	return "merchant_stores"
}
