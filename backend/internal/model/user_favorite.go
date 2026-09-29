package model

import "time"

// UserFavorite stores a user's favorite products for a merchant and store.
type UserFavorite struct {
	MerchantID string    `json:"-" gorm:"primaryKey;size:64"`
	StoreID    string    `json:"storeId" gorm:"primaryKey;size:64"`
	AppID      string    `json:"-" gorm:"primaryKey;size:64"`
	OpenID     string    `json:"-" gorm:"primaryKey;size:128"`
	ProductID  string    `json:"productId" gorm:"primaryKey;size:128"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (UserFavorite) TableName() string {
	return "user_favorites"
}
