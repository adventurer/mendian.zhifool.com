package model

// Menu 表示单个商户的菜单聚合，系列通过关联表保存。
type Menu struct {
	MerchantID string       `json:"merchantId" gorm:"primaryKey;size:64"`
	BrandName  string       `json:"brandName" gorm:"size:128;not null"`
	Series     []MenuSeries `json:"series" gorm:"foreignKey:MerchantID;references:MerchantID"`
}

// TableName 指定菜单数据在数据库中的表名。
func (Menu) TableName() string {
	return "menus"
}

// MenuSeries 表示菜单中的一个商品系列及其商品列表。
type MenuSeries struct {
	MerchantID string    `json:"merchantId" gorm:"primaryKey;size:64"`
	ID         string    `json:"id" gorm:"primaryKey;size:128"`
	Name       string    `json:"name" gorm:"size:128;not null"`
	Icon       string    `json:"icon" gorm:"size:255"`
	Badge      string    `json:"badge,omitempty" gorm:"size:32"`
	SortOrder  int       `json:"-" gorm:"not null;default:0"`
	Products   []Product `json:"products" gorm:"foreignKey:MerchantID,SeriesID;references:MerchantID,ID"`
}

// Product 表示可售商品及其价格、图片和可选规格。
type Product struct {
	MerchantID  string          `json:"merchantId" gorm:"primaryKey;size:64"`
	ID          string          `json:"id" gorm:"primaryKey;size:128"`
	SeriesID    string          `json:"-" gorm:"size:128;not null;index"`
	Name        string          `json:"name" gorm:"size:255;not null"`
	Description string          `json:"description" gorm:"type:text"`
	Price       float64         `json:"price" gorm:"not null"`
	Image       string          `json:"image" gorm:"size:255"`
	SortOrder   int             `json:"-" gorm:"not null;default:0"`
	Options     []ProductOption `json:"options" gorm:"foreignKey:MerchantID,ProductID;references:MerchantID,ID"`
}

// ProductOption 表示商品的一组选项，例如杯型、奶类或甜度。
type ProductOption struct {
	MerchantID    string               `json:"-" gorm:"primaryKey;size:64"`
	ProductID     string               `json:"-" gorm:"primaryKey;size:128"`
	ID            string               `json:"id" gorm:"primaryKey;size:128"`
	Title         string               `json:"title" gorm:"size:128;not null"`
	SelectedIndex int                  `json:"selectedIndex" gorm:"not null;default:0"`
	SortOrder     int                  `json:"-" gorm:"not null;default:0"`
	Values        []ProductOptionValue `json:"values" gorm:"foreignKey:MerchantID,ProductID,OptionID;references:MerchantID,ProductID,ID"`
}

// ProductOptionValue 表示商品规格中的一个可选值及其价格增量。
type ProductOptionValue struct {
	ID         uint    `json:"-" gorm:"primaryKey;autoIncrement"`
	MerchantID string  `json:"-" gorm:"size:64;not null;uniqueIndex:idx_product_option_value_order,priority:1"`
	ProductID  string  `json:"-" gorm:"size:128;not null;uniqueIndex:idx_product_option_value_order,priority:2"`
	OptionID   string  `json:"-" gorm:"size:128;not null;uniqueIndex:idx_product_option_value_order,priority:3"`
	SortOrder  int     `json:"-" gorm:"not null;uniqueIndex:idx_product_option_value_order,priority:4"`
	Label      string  `json:"label" gorm:"size:255;not null"`
	ExtraPrice float64 `json:"extraPrice" gorm:"not null;default:0"`
}

func (MenuSeries) TableName() string {
	return "menu_series"
}

func (Product) TableName() string {
	return "products"
}

func (ProductOption) TableName() string {
	return "product_options"
}

func (ProductOptionValue) TableName() string {
	return "product_option_values"
}

// CartItem 表示购物车中的一个商品条目及其所选规格和单价。
type CartItem struct {
	ID         uint    `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	MerchantID string  `json:"merchantId" gorm:"size:64;not null;index:idx_cart_items_scope,priority:1;index:idx_cart_items_store_cart,priority:1"`
	StoreID    string  `json:"storeId" gorm:"size:64;not null;default:default;index:idx_cart_items_store_cart,priority:2"`
	CartID     string  `json:"cartId" gorm:"size:64;not null;index:idx_cart_items_scope,priority:2;index:idx_cart_items_store_cart,priority:3"`
	AppID      string  `json:"-" gorm:"size:64;not null;default:'';index:idx_cart_items_owner,priority:1"`
	OpenID     string  `json:"-" gorm:"size:128;not null;default:'';index:idx_cart_items_owner,priority:2"`
	ProductID  string  `json:"productId" gorm:"size:128;not null;index"`
	Name       string  `json:"name" gorm:"size:255;not null"`
	Quantity   int     `json:"quantity" gorm:"not null"`
	UnitPrice  float64 `json:"unitPrice" gorm:"not null"`
	Options    string  `json:"options" gorm:"type:text"`
}

// TableName 指定购物车条目在数据库中的表名。
func (CartItem) TableName() string {
	return "cart_items"
}
