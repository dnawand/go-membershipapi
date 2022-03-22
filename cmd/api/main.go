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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed swagger
var swagger embed.FS

func main() {
	logger := zapConfig()
	defer logger.Sync()

	dbConfig, err := dbConfig()
	if err != nil {
		log.Fatal(err)
	}

	userRepository := repositories.NewUserRepository(dbConfig)
	productRepository := repositories.NewProductRepository(dbConfig)
	subscriptionRespository := repositories.NewSubscriptionRepository(dbConfig)

	userService := app.NewUserService(userRepository)
	productService := app.NewProductService(productRepository)
	voucherStorage := loadVouchers()
	subscriptionService := app.NewSubscriptionService(subscriptionRespository, userRepository, productRepository, voucherStorage)

	userHandler := handlers.NewUserHandler(logger, userService)
	productHandler := handlers.NewProductHandler(logger, productService)
	subscriptionHandler := handlers.NewSubscriptionHandler(logger, subscriptionService)

	router := gin.Default()

	router.POST("/users", userHandler.Create)
	router.POST("/products", productHandler.Create)
	router.GET("/products/:product-id", productHandler.Fetch)
	router.GET("/products", productHandler.List)
	router.POST("/subscriptions", subscriptionHandler.Create)
	router.GET("/subscriptions/:subscription-id", subscriptionHandler.Fetch)
	router.GET("/users/:user-id/subscriptions", subscriptionHandler.List)
	router.PATCH("/subscriptions/:subscription-id", subscriptionHandler.Action)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	fileRouter := mux.NewRouter()
	fileRouter.PathPrefix("/swagger").Handler(http.FileServer(http.FS(swagger))).Methods(http.MethodGet)

	fileServer := &http.Server{
		Addr:         ":8081",
		Handler:      fileRouter,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	ok := gracefulRun(server, fileServer, logger)
	if !ok {
		logger.Info("server forced to shutdown")
		logger.Sync()
		log.Fatal("server exiting")
	}

	logger.Info("server exiting")
}

func zapConfig() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("could not initialize zao logger: %v", err)
	}

	return logger
}

func dbConfig() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=secretpw dbname=subscription port=5432 sslmode=disable"
	cfg := &gorm.Config{}
	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, fmt.Errorf("could not get db config: %w", err)
	}

	err = db.AutoMigrate(
		domain.User{},
		domain.Plan{},
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

	id, _ := uuid.NewRandom()
	voucherFixedAmount := domain.Voucher{
		ID:       id.String(),
		Type:     domain.VoucherFixedAmount,
		Discount: 5,
		IsActive: true,
	}

	id, _ = uuid.NewRandom()
	voucherPercentage := domain.Voucher{
		ID:       id.String(),
		Type:     domain.VoucherPercentage,
		Discount: 10,
		IsActive: true,
	}

	id, _ = uuid.NewRandom()
	voucherInactive := domain.Voucher{
		ID:       id.String(),
		Type:     domain.VoucherPercentage,
		Discount: 10,
		IsActive: false,
	}

	voucherStorage.Save(voucherFixedAmount.ID, voucherFixedAmount)
	voucherStorage.Save(voucherPercentage.ID, voucherPercentage)
	voucherStorage.Save(voucherInactive.ID, voucherInactive)

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

	quit := make(chan os.Signal)

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
