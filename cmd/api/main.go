package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnawand/go-membershipapi/internal/handlers"
	"github.com/dnawand/go-membershipapi/internal/storage"
	"github.com/dnawand/go-membershipapi/pkg/app"
	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/dnawand/go-membershipapi/pkg/repositories"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed swagger
var swagger embed.FS

func main() {
	logger := zapConfig()
	defer logger.Sync()

	dbConfig, err := dbConfig()
	if err != nil {
		logger.Error("could not initialize database configuration", zap.Error(err))
		logger.Sync()
		os.Exit(1)
	}

	voucherStorage := loadVouchers()
	userRepository := repositories.NewUserRepository(dbConfig)
	productRepository := repositories.NewProductRepository(dbConfig)
	subscriptionRespository := repositories.NewSubscriptionRepository(dbConfig, voucherStorage)

	userService := app.NewUserService(userRepository)
	productService := app.NewProductService(productRepository)
	discountService := app.NewDiscountService()
	subscriptionService := app.NewSubscriptionService(
		subscriptionRespository, userRepository, productRepository, voucherStorage, discountService,
	)

	userHandler := handlers.NewUserHandler(logger, userService)
	productHandler := handlers.NewProductHandler(logger, productService)
	subscriptionHandler := handlers.NewSubscriptionHandler(logger, subscriptionService)

	router := configRouter(userHandler, productHandler, subscriptionHandler)
	server, fileServer := serverConfig(router)
	ok := gracefulRun(server, fileServer, logger)
	if !ok {
		logger.Info("server forced to shutdown")
		logger.Sync()
		os.Exit(1)
	}

	logger.Info("server exiting")
}

func configRouter(
	userHandler *handlers.UserHandler,
	productHandler *handlers.ProductHandler,
	subscriptionHandler *handlers.SubscriptionHandler,
) *gin.Engine {
	router := gin.Default()

	router.POST("/users", userHandler.Create)
	router.GET("/users/:user-id", userHandler.Fetch)
	router.POST("/products", productHandler.Create)
	router.GET("/products/:product-id", productHandler.Fetch)
	router.GET("/products", productHandler.List)
	router.POST("/users/:user-id/subscriptions", subscriptionHandler.Create)
	router.GET("/users/:user-id/subscriptions/:subscription-id", subscriptionHandler.Fetch)
	router.GET("/users/:user-id/subscriptions", subscriptionHandler.List)
	router.PATCH("/users/:user-id/subscriptions/:subscription-id", subscriptionHandler.Action)

	return router
}

func serverConfig(router *gin.Engine) (server *http.Server, fileServer *http.Server) {
	server = &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	mux := http.NewServeMux()
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.FS(swagger))))

	fileServer = &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	return server, fileServer
}

func zapConfig() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("could not initialize zao logger: %v", err)
	}

	return logger
}

func dbConfig() (*gorm.DB, error) {
	cfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), cfg)

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		domain.User{},
		domain.ProductPlan{},
		domain.SubscriptionPlan{},
		domain.Product{},
		domain.Subscription{},
	)
	if err != nil {
		return nil, fmt.Errorf("could not migrate models: %w", err)
	}

	return db, err
}

func loadVouchers() *storage.Store {
	voucherStorage := storage.NewStore()

	voucherFixedAmount := domain.Voucher{
		ID:       "b86b4903-2043-4f71-b154-efec19fbc55a",
		Type:     domain.VoucherFixedAmount,
		Discount: "5.00",
		IsActive: true,
	}
	voucherPercentage := domain.Voucher{
		ID:       "4976ff21-a188-4bcc-97a0-2cf2278e9a6b",
		Type:     domain.VoucherPercentage,
		Discount: "10.10",
		IsActive: true,
	}
	voucherInactive := domain.Voucher{
		ID:       "18c4b4ea-6fce-4ee7-8d3b-a16047a8789e",
		Type:     domain.VoucherPercentage,
		Discount: "10.10",
		IsActive: false,
	}

	voucherStorage.Save(voucherFixedAmount.ID, voucherFixedAmount)
	voucherStorage.Save(voucherPercentage.ID, voucherPercentage)
	voucherStorage.Save(voucherInactive.ID, voucherInactive)

	fmt.Println("============================ Vouchers ============================")
	fmt.Printf("FixedAmount: %s Discount: %s\n", voucherFixedAmount.ID, voucherFixedAmount.Discount)
	fmt.Printf("Percentage: %s Discount: %s\n", voucherPercentage.ID, voucherPercentage.Discount)
	fmt.Printf("Inactive: %s Discount: %s\n", voucherInactive.ID, voucherInactive.Discount)
	fmt.Println("==================================================================")

	return voucherStorage
}

func gracefulRun(server *http.Server, fileServer *http.Server, logger *zap.Logger) (ok bool) {
	ok = true

	go func() {
		logger.Info(fmt.Sprintf("starting server on %s", server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("could not start the server", zap.Error(err))
		}
	}()

	go func() {
		logger.Info(fmt.Sprintf("starting file server on %s", fileServer.Addr))

		if err := fileServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("could not start the file server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("server shutting down...")

	errFileServer := fileServer.Close()
	if errFileServer != nil {
		logger.Error("erro when closing file server", zap.Error(errFileServer))
		ok = false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server need to be forced to shutdown", zap.Error(err))
		ok = false
	}

	return ok
}
