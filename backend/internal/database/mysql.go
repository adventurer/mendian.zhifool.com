package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"mendian-backend/internal/config"
	"mendian-backend/internal/model"
	"mendian-backend/internal/seed"
)

func CreateDatabase(cfg config.DatabaseConfig) error {
	statement, err := createDatabaseStatement(cfg.Name, cfg.Charset)
	if err != nil {
		return err
	}

	serverConfig := cfg
	serverConfig.Name = ""
	dsn, err := buildDSN(serverConfig)
	if err != nil {
		return err
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql server connection: %w", err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to mysql server: %w", err)
	}
	if _, err := sqlDB.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("create mysql database: %w", err)
	}
	return nil
}

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: false,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}
	legacyMenus, hasLegacyTables, err := readExistingMenus(db)
	if err != nil {
		return nil, fmt.Errorf("read legacy menu data: %w", err)
	}
	if err := db.AutoMigrate(
		&model.Merchant{},
		&model.Store{},
		&model.StoreMenu{},
		&model.StoreMenuSeries{},
		&model.StoreProduct{},
		&model.StoreProductOption{},
		&model.StoreProductOptionValue{},
		&model.CartItem{},
		&model.PaymentOrder{},
		&model.WeChatUser{},
		&model.UserAddress{},
	); err != nil {
		return nil, fmt.Errorf("migrate mysql schema: %w", err)
	}
	if hasLegacyTables {
		if err := migrateLegacyMenusToStores(db, legacyMenus); err != nil {
			return nil, fmt.Errorf("migrate legacy menus to stores: %w", err)
		}
		if err := dropLegacyMenuTables(db); err != nil {
			return nil, fmt.Errorf("drop legacy menu tables: %w", err)
		}
	}

	return db, nil
}

func readExistingMenus(db *gorm.DB) ([]model.Menu, bool, error) {
	if !db.Migrator().HasTable("menus") {
		return nil, false, nil
	}
	if db.Migrator().HasColumn("menus", "series") {
		return readLegacyMenus(db)
	}
	for _, table := range []string{"menu_series", "products", "product_options", "product_option_values"} {
		if !db.Migrator().HasTable(table) {
			return nil, true, fmt.Errorf("legacy menu table %q is missing", table)
		}
	}
	ordered := func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }
	var menus []model.Menu
	if err := db.
		Preload("Series", ordered).
		Preload("Series.Products", ordered).
		Preload("Series.Products.Options", ordered).
		Preload("Series.Products.Options.Values", ordered).
		Find(&menus).Error; err != nil {
		return nil, true, err
	}
	return menus, true, nil
}

func migrateLegacyMenusToStores(db *gorm.DB, menus []model.Menu) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, menu := range menus {
			storeName := menu.BrandName
			address := ""
			latitude, longitude := float64(0), float64(0)
			if menu.MerchantID == seed.DefaultMerchantID {
				storeName = "贵阳荟华里店"
				address = "贵州省贵阳市"
				latitude, longitude = 26.6477, 106.6302
			}
			store := model.Store{
				MerchantID: menu.MerchantID,
				StoreID:    model.DefaultStoreID,
				Name:       storeName,
				Address:    address,
				Latitude:   latitude,
				Longitude:  longitude,
				IsActive:   true,
				IsDefault:  true,
			}
			if err := tx.Where("merchant_id = ? AND store_id = ?", menu.MerchantID, model.DefaultStoreID).FirstOrCreate(&store).Error; err != nil {
				return err
			}
			if err := seed.ReplaceStoreMenu(tx, model.DefaultStoreID, menu); err != nil {
				return err
			}
		}
		return nil
	})
}

func dropLegacyMenuTables(db *gorm.DB) error {
	for _, table := range []string{"product_option_values", "product_options", "products", "menu_series", "menus"} {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return fmt.Errorf("drop table %s: %w", table, err)
			}
		}
	}
	return nil
}

func readLegacyMenus(db *gorm.DB) ([]model.Menu, bool, error) {
	if !db.Migrator().HasTable("menus") || !db.Migrator().HasColumn("menus", "series") {
		return nil, false, nil
	}

	rows, err := db.Raw("SELECT merchant_id, brand_name, `series` FROM `menus`").Rows()
	if err != nil {
		return nil, true, err
	}
	defer rows.Close()

	menus := make([]model.Menu, 0)
	for rows.Next() {
		var menu model.Menu
		var seriesJSON []byte
		if err := rows.Scan(&menu.MerchantID, &menu.BrandName, &seriesJSON); err != nil {
			return nil, true, err
		}
		if len(seriesJSON) > 0 {
			if err := json.Unmarshal(seriesJSON, &menu.Series); err != nil {
				return nil, true, fmt.Errorf("decode menu %q series: %w", menu.MerchantID, err)
			}
		}
		for seriesIndex := range menu.Series {
			menu.Series[seriesIndex].MerchantID = menu.MerchantID
			for productIndex := range menu.Series[seriesIndex].Products {
				menu.Series[seriesIndex].Products[productIndex].MerchantID = menu.MerchantID
			}
		}
		menus = append(menus, menu)
	}
	if err := rows.Err(); err != nil {
		return nil, true, err
	}
	return menus, true, nil
}

func createDatabaseStatement(name, charset string) (string, error) {
	if name == "" || len(name) > 64 {
		return "", fmt.Errorf("database name must contain 1 to 64 characters")
	}
	for _, char := range charset {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return "", fmt.Errorf("invalid database charset %q", charset)
		}
	}
	if charset == "" {
		return "", fmt.Errorf("database charset is required")
	}

	quotedName := strings.ReplaceAll(name, "`", "``")
	return fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s", quotedName, charset), nil
}

func buildDSN(cfg config.DatabaseConfig) (string, error) {
	location := time.Local
	if cfg.Location != "" {
		var err error
		location, err = time.LoadLocation(cfg.Location)
		if err != nil {
			return "", fmt.Errorf("load database location: %w", err)
		}
	}

	dsn := mysqlDriver.Config{
		User:                 cfg.Username,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		DBName:               cfg.Name,
		Params:               map[string]string{"charset": cfg.Charset},
		ParseTime:            cfg.ParseTime,
		Loc:                  location,
		AllowNativePasswords: true,
	}
	return dsn.FormatDSN(), nil
}
