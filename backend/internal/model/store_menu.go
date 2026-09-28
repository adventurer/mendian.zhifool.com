package model

// StoreMenu holds a store's independently configurable catalog.
type StoreMenu struct {
	MerchantID string            `json:"merchantId" gorm:"primaryKey;size:64"`
	StoreID    string            `json:"storeId" gorm:"primaryKey;size:64"`
	BrandName  string            `json:"brandName" gorm:"size:128;not null"`
	Series     []StoreMenuSeries `json:"series" gorm:"foreignKey:MerchantID,StoreID;references:MerchantID,StoreID"`
}

func (StoreMenu) TableName() string {
	return "store_menus"
}

type StoreMenuSeries struct {
	MerchantID string         `json:"merchantId" gorm:"primaryKey;size:64"`
	StoreID    string         `json:"storeId" gorm:"primaryKey;size:64"`
	ID         string         `json:"id" gorm:"primaryKey;size:128"`
	Name       string         `json:"name" gorm:"size:128;not null"`
	Icon       string         `json:"icon" gorm:"size:255"`
	Badge      string         `json:"badge,omitempty" gorm:"size:32"`
	SortOrder  int            `json:"-" gorm:"not null;default:0"`
	Products   []StoreProduct `json:"products" gorm:"foreignKey:MerchantID,StoreID,SeriesID;references:MerchantID,StoreID,ID"`
}

func (StoreMenuSeries) TableName() string {
	return "store_menu_series"
}

type StoreProduct struct {
	MerchantID  string               `json:"merchantId" gorm:"primaryKey;size:64"`
	StoreID     string               `json:"storeId" gorm:"primaryKey;size:64"`
	ID          string               `json:"id" gorm:"primaryKey;size:128"`
	SeriesID    string               `json:"seriesId" gorm:"size:128;not null;index"`
	Name        string               `json:"name" gorm:"size:255;not null"`
	Description string               `json:"description" gorm:"type:text"`
	Price       float64              `json:"price" gorm:"not null"`
	Image       string               `json:"image" gorm:"size:255"`
	SortOrder   int                  `json:"-" gorm:"not null;default:0"`
	Options     []StoreProductOption `json:"options" gorm:"foreignKey:MerchantID,StoreID,ProductID;references:MerchantID,StoreID,ID"`
}

func (StoreProduct) TableName() string {
	return "store_products"
}

type StoreProductOption struct {
	MerchantID    string                    `json:"merchantId" gorm:"primaryKey;size:64"`
	StoreID       string                    `json:"storeId" gorm:"primaryKey;size:64"`
	ProductID     string                    `json:"productId" gorm:"primaryKey;size:128"`
	ID            string                    `json:"id" gorm:"primaryKey;size:128"`
	Title         string                    `json:"title" gorm:"size:128;not null"`
	SelectedIndex int                       `json:"selectedIndex" gorm:"not null;default:0"`
	SortOrder     int                       `json:"-" gorm:"not null;default:0"`
	Values        []StoreProductOptionValue `json:"values" gorm:"foreignKey:MerchantID,StoreID,ProductID,OptionID;references:MerchantID,StoreID,ProductID,ID"`
}

func (StoreProductOption) TableName() string {
	return "store_product_options"
}

type StoreProductOptionValue struct {
	ID         uint    `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	MerchantID string  `json:"merchantId" gorm:"size:64;not null;uniqueIndex:idx_store_product_option_value_order,priority:1"`
	StoreID    string  `json:"storeId" gorm:"size:64;not null;uniqueIndex:idx_store_product_option_value_order,priority:2"`
	ProductID  string  `json:"productId" gorm:"size:128;not null;uniqueIndex:idx_store_product_option_value_order,priority:3"`
	OptionID   string  `json:"optionId" gorm:"size:128;not null;uniqueIndex:idx_store_product_option_value_order,priority:4"`
	SortOrder  int     `json:"sortOrder" gorm:"not null;uniqueIndex:idx_store_product_option_value_order,priority:5"`
	Label      string  `json:"label" gorm:"size:255;not null"`
	ExtraPrice float64 `json:"extraPrice" gorm:"not null;default:0"`
}

func (StoreProductOptionValue) TableName() string {
	return "store_product_option_values"
}
