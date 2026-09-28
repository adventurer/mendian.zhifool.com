package seed

import (
	"gorm.io/gorm"

	"mendian-backend/internal/model"
)

const DefaultMerchantID = "demo-merchant"

func EnsureDefaultMenu(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		merchant := model.Merchant{MerchantID: DefaultMerchantID, Name: "知甜蛋糕"}
		if err := tx.Where("merchant_id = ?", DefaultMerchantID).FirstOrCreate(&merchant).Error; err != nil {
			return err
		}
		store := model.Store{
			MerchantID: DefaultMerchantID,
			StoreID:    model.DefaultStoreID,
			Name:       "贵阳荟华里店",
			Address:    "贵州省贵阳市",
			Latitude:   26.6477,
			Longitude:  106.6302,
			IsActive:   true,
			IsDefault:  true,
		}
		if err := tx.Where("merchant_id = ? AND store_id = ?", DefaultMerchantID, model.DefaultStoreID).FirstOrCreate(&store).Error; err != nil {
			return err
		}
		defaultMenu := DefaultMenu()
		if err := PersistStoreMenu(tx, model.DefaultStoreID, defaultMenu); err != nil {
			return err
		}

		demoStore := model.Store{
			MerchantID: DefaultMerchantID,
			StoreID:    "demo-store",
			Name:       "多门店演示店（测试）",
			Address:    "演示数据，实际地址待配置",
			IsActive:   true,
			IsDefault:  false,
			IsTest:     true,
		}
		if err := tx.Where("merchant_id = ? AND store_id = ?", DefaultMerchantID, demoStore.StoreID).FirstOrCreate(&demoStore).Error; err != nil {
			return err
		}
		return PersistStoreMenu(tx, demoStore.StoreID, DemoStoreMenu())
	})
}

func PersistStoreMenu(tx *gorm.DB, storeID string, menu model.Menu) error {
	storeMenu := model.StoreMenu{MerchantID: menu.MerchantID, StoreID: storeID, BrandName: menu.BrandName}
	if err := tx.Where("merchant_id = ? AND store_id = ?", menu.MerchantID, storeID).FirstOrCreate(&storeMenu).Error; err != nil {
		return err
	}
	for seriesIndex, series := range menu.Series {
		seriesRecord := model.StoreMenuSeries{
			MerchantID: menu.MerchantID,
			StoreID:    storeID,
			ID:         series.ID,
			Name:       series.Name,
			Icon:       series.Icon,
			Badge:      series.Badge,
			SortOrder:  seriesIndex,
		}
		if err := tx.Where("merchant_id = ? AND store_id = ? AND id = ?", menu.MerchantID, storeID, series.ID).FirstOrCreate(&seriesRecord).Error; err != nil {
			return err
		}
		for productIndex, product := range series.Products {
			productRecord := model.StoreProduct{
				MerchantID:  menu.MerchantID,
				StoreID:     storeID,
				ID:          product.ID,
				SeriesID:    series.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Image:       product.Image,
				SortOrder:   productIndex,
			}
			if err := tx.Where("merchant_id = ? AND store_id = ? AND id = ?", menu.MerchantID, storeID, product.ID).FirstOrCreate(&productRecord).Error; err != nil {
				return err
			}
			for optionIndex, option := range product.Options {
				optionRecord := model.StoreProductOption{
					MerchantID:    menu.MerchantID,
					StoreID:       storeID,
					ProductID:     product.ID,
					ID:            option.ID,
					Title:         option.Title,
					SelectedIndex: option.SelectedIndex,
					SortOrder:     optionIndex,
				}
				if err := tx.Where("merchant_id = ? AND store_id = ? AND product_id = ? AND id = ?", menu.MerchantID, storeID, product.ID, option.ID).FirstOrCreate(&optionRecord).Error; err != nil {
					return err
				}
				for valueIndex, value := range option.Values {
					valueRecord := model.StoreProductOptionValue{
						MerchantID: menu.MerchantID,
						StoreID:    storeID,
						ProductID:  product.ID,
						OptionID:   option.ID,
						SortOrder:  valueIndex,
						Label:      value.Label,
						ExtraPrice: value.ExtraPrice,
					}
					if err := tx.Where("merchant_id = ? AND store_id = ? AND product_id = ? AND option_id = ? AND sort_order = ?", menu.MerchantID, storeID, product.ID, option.ID, valueIndex).FirstOrCreate(&valueRecord).Error; err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func ReplaceStoreMenu(tx *gorm.DB, storeID string, menu model.Menu) error {
	conditions := map[string]any{"merchant_id": menu.MerchantID, "store_id": storeID}
	for _, target := range []any{
		&model.StoreProductOptionValue{},
		&model.StoreProductOption{},
		&model.StoreProduct{},
		&model.StoreMenuSeries{},
		&model.StoreMenu{},
	} {
		if err := tx.Where(conditions).Delete(target).Error; err != nil {
			return err
		}
	}
	return PersistStoreMenu(tx, storeID, menu)
}

func DefaultMenu() model.Menu {
	return model.Menu{
		MerchantID: DefaultMerchantID,
		BrandName:  "知甜蛋糕",
		Series: []model.MenuSeries{
			series("seasonal", "季节新品", "icons/sparkles.svg",
				product("rose-rice-latte", "玫瑰米酿拿铁", "玫瑰芬芳融入温润发酵米香，伴随淡淡桂花清香与苹果甜香。含少量酒精（低于0.5%vol），孕妇、驾驶人士及未成年人请谨慎选择。", 20, "coffee/rose-latte.jpg"),
				product("butter-soe-latte", "黄油SOE拿铁", "精选黄油棕果SOE咖啡豆，不额外加糖，烘烤榛果温润扎实。", 20, "coffee/butter-latte.jpg"),
				product("osmanthus-iced-latte", "桂花冰拿铁", "浓郁咖啡与清甜桂花，入口清爽柔和。", 22, "coffee/osmanthus-latte.jpg")),
			series("reserve", "珍藏系列", "icons/coffee.svg", product("reserve-latte", "珍藏拿铁", "甄选咖啡豆搭配醇厚牛奶，香气饱满顺滑。", 22, "coffee/butter-latte.jpg")),
			seriesWithBadge("fruit-americano", "果味美式", "icons/citrus.svg", "爆", product("citrus-americano", "柑橘美式", "清新柑橘香气融入醇正美式，酸甜清爽。", 20, "coffee/osmanthus-latte.jpg")),
			series("large-cup", "超大杯系列", "icons/cup-soda.svg", product("large-americano", "超大杯美式", "双份浓缩搭配清冽冷水，满足大杯畅饮。", 18, "coffee/butter-latte.jpg")),
			series("classic-espresso", "经典意式", "icons/coffee.svg", product("classic-latte", "经典拿铁", "浓缩咖啡与绵密牛奶融合，口感平衡柔和。", 20, "coffee/rose-latte.jpg")),
			seriesWithBadge("soe", "单品豆SOE", "icons/bean.svg", "新", product("soe-americano", "单品豆SOE美式", "单一产地咖啡豆呈现明亮果香与干净回甘。", 18, "coffee/osmanthus-latte.jpg")),
			series("milk-coffee", "热卖奶咖", "icons/milk.svg", product("hot-milk-latte", "热卖奶咖", "现萃浓缩与热牛奶相融，口感醇厚细腻。", 20, "coffee/butter-latte.jpg")),
			series("oat", "燕麦系列", "icons/wheat.svg", product("oat-latte", "燕麦拿铁", "植物燕麦奶带来谷物香气，轻盈顺口。", 23, "coffee/rose-latte.jpg")),
			series("classic-americano", "经典美式", "icons/coffee.svg",
				product("classic-americano-ice", "经典冰美式", "双份浓缩融合清冽冰水，口感干净明亮。", 16, "coffee/butter-latte.jpg"),
				product("classic-americano-hot", "经典热美式", "醇厚浓缩与热水相融，带来平衡顺口的风味。", 16, "coffee/rose-latte.jpg")),
			seriesWithBadge("flavored-latte", "风味拿铁", "icons/milk.svg", "新",
				product("vanilla-latte", "香草拿铁", "香草甜香与浓缩咖啡交织，口感柔和醇厚。", 24, "coffee/rose-latte.jpg"),
				product("caramel-latte", "焦糖拿铁", "焦糖香气融入绵密牛奶，甜润而不腻。", 24, "coffee/butter-latte.jpg")),
			series("tea-coffee", "茶咖特调", "icons/citrus.svg",
				product("osmanthus-tea-coffee", "桂花茶咖", "桂花清香与咖啡醇香相遇，层次清新。", 23, "coffee/osmanthus-latte.jpg"),
				product("citrus-tea-coffee", "柑橘冷萃", "柑橘果香融入冷萃咖啡，酸甜爽口。", 22, "coffee/osmanthus-latte.jpg")),
			seriesWithBadge("espresso-special", "浓缩特调", "icons/bean.svg", "新",
				product("espresso-tonic", "浓缩汤力", "浓缩咖啡搭配气泡汤力水，清爽带有柑橘香。", 22, "coffee/osmanthus-latte.jpg"),
				product("dirty-coffee", "脏脏咖啡", "热浓缩缓缓注入冰牛奶，呈现浓郁冷热交融。", 24, "coffee/butter-latte.jpg")),
			series("payment-test", "支付测试", "icons/coffee.svg", model.Product{
				MerchantID:  DefaultMerchantID,
				ID:          "one-fen-payment-test",
				SeriesID:    "payment-test",
				Name:        "支付测试商品（1分）",
				Description: "用于微信支付流程验证，单价为人民币0.01元。",
				Price:       0.01,
				Image:       "coffee/rose-latte.jpg",
				Options: []model.ProductOption{{
					ID:            "test-option",
					Title:         "商品",
					SelectedIndex: 0,
					Values:        []model.ProductOptionValue{{Label: "支付测试商品", ExtraPrice: 0}},
				}},
			}),
		},
	}
}

func DemoStoreMenu() model.Menu {
	return model.Menu{
		MerchantID: DefaultMerchantID,
		BrandName:  "知甜蛋糕",
		Series: []model.MenuSeries{{
			MerchantID: DefaultMerchantID,
			ID:         "demo-products",
			Name:       "演示商品",
			Icon:       "icons/sparkles.svg",
			Products: []model.Product{
				{MerchantID: DefaultMerchantID, ID: "demo-latte", Name: "演示拿铁（测试）", Description: "测试商品：零元演示数据，不可支付。", Price: 0, Image: "coffee/rose-latte.jpg", Options: []model.ProductOption{}},
				{MerchantID: DefaultMerchantID, ID: "demo-americano", Name: "演示美式（测试）", Description: "测试商品：零元演示数据，不可支付。", Price: 0, Image: "coffee/butter-latte.jpg", Options: []model.ProductOption{}},
				{MerchantID: DefaultMerchantID, ID: "demo-citrus", Name: "演示柑橘饮（测试）", Description: "测试商品：零元演示数据，不可支付。", Price: 0, Image: "coffee/osmanthus-latte.jpg", Options: []model.ProductOption{}},
			},
		}},
	}
}

func series(id, name, icon string, products ...model.Product) model.MenuSeries {
	return model.MenuSeries{MerchantID: DefaultMerchantID, ID: id, Name: name, Icon: icon, Products: products}
}

func seriesWithBadge(id, name, icon, badge string, products ...model.Product) model.MenuSeries {
	item := series(id, name, icon, products...)
	item.Badge = badge
	return item
}

func product(id, name, description string, price float64, image string) model.Product {
	return model.Product{
		MerchantID:  DefaultMerchantID,
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		Image:       image,
		Options: []model.ProductOption{
			{
				ID: "size", Title: "杯型及温度", SelectedIndex: 0,
				Values: []model.ProductOptionValue{
					{Label: "超大冰杯 473ml（正常冰）", ExtraPrice: 5},
					{Label: "超大热杯 473ml", ExtraPrice: 5},
					{Label: "大冰杯 355ml（正常冰）", ExtraPrice: 0},
					{Label: "大热杯 355ml", ExtraPrice: 0},
				},
			},
			{
				ID: "milk", Title: "奶搭配", SelectedIndex: 0,
				Values: []model.ProductOptionValue{{Label: "牛奶", ExtraPrice: 0}, {Label: "燕麦奶", ExtraPrice: 3}},
			},
			{
				ID: "sweetness", Title: "甜度", SelectedIndex: 0,
				Values: []model.ProductOptionValue{{Label: "标准糖", ExtraPrice: 0}, {Label: "少糖", ExtraPrice: 0}, {Label: "不另外加糖", ExtraPrice: 0}},
			},
		},
	}
}
