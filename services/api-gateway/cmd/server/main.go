package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"paylater/services/api-gateway/internal/config"
	"paylater/services/api-gateway/internal/proxy"
	"paylater/services/api-gateway/internal/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	proxies := routes.Proxies{
		Auth:     proxy.New("auth", cfg.AuthServiceURL, cfg.ProxyTimeout),
		User:     proxy.New("user", cfg.UserServiceURL, cfg.ProxyTimeout),
		Merchant: proxy.New("merchant", cfg.MerchantServiceURL, cfg.ProxyTimeout),
		Ledger:   proxy.New("ledger", cfg.LedgerServiceURL, cfg.ProxyTimeout),
		Report:   proxy.New("report", cfg.ReportServiceURL, cfg.ProxyTimeout),
	}

	router := gin.New()
	router.Use(gin.Recovery())
	routes.Setup(router, proxies)

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	log.Printf("api-gateway listening on :%s", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
