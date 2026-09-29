package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/kataras/iris/v12"

	"mendian-backend/internal/api"
	"mendian-backend/internal/config"
	"mendian-backend/internal/database"
	"mendian-backend/internal/payment"
	"mendian-backend/internal/seed"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	if err := seed.EnsureDefaultMenu(db); err != nil {
		log.Fatalf("seed default merchant menu: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}
	defer sqlDB.Close()

	app := iris.New()
	app.UseRouter(func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")
		if ctx.Method() == http.MethodOptions {
			ctx.StatusCode(iris.StatusNoContent)
			return
		}
		ctx.Next()
	})
	app.HandleDir("/assets", iris.Dir("./assets"), iris.DirOptions{ShowList: false, ShowHidden: false})
	app.Get("/api/health", func(ctx iris.Context) {
		if err := sqlDB.PingContext(ctx.Request().Context()); err != nil {
			ctx.StatusCode(iris.StatusServiceUnavailable)
			ctx.JSON(iris.Map{"status": "unavailable"})
			return
		}
		ctx.JSON(iris.Map{"status": "ok"})
	})
	api.RegisterRoutes(app, db)
	paymentGateways := make(map[string]*payment.Gateway)
	for _, payConfig := range cfg.WeChatPay.MerchantConfigs() {
		paymentGateway, gatewayErr := payment.NewGateway(context.Background(), payConfig)
		if gatewayErr != nil {
			log.Printf("WeChat Pay unavailable for merchant %q: %v", payConfig.MerchantID, gatewayErr)
			continue
		}
		paymentGateways[payConfig.MerchantID] = paymentGateway
	}
	if len(paymentGateways) == 0 && len(cfg.WeChatPay.MerchantConfigs()) > 0 {
		log.Printf("WeChat Pay unavailable for all configured merchants")
	}
	api.RegisterPaymentRoutesForMerchants(app, db, paymentGateways)

	address := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	runner := iris.Addr(address)
	if cfg.Server.AutoTLSDomain != "" {
		runner = iris.AutoTLS(":443", cfg.Server.AutoTLSDomain, cfg.Server.AutoTLSEmail)
	} else if cfg.Server.TLSCert != "" {
		runner = iris.TLS(address, cfg.Server.TLSCert, cfg.Server.TLSKey)
	}
	if err := app.Run(runner); err != nil {
		log.Fatal(err)
	}
}
