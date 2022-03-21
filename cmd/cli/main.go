package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/dnawand/go-subscriptionapi/internal/handlers"
	"github.com/dnawand/go-subscriptionapi/pkg/app"
	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/dnawand/go-subscriptionapi/pkg/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed swagger/*
var swagger embed.FS

func main() {
	dbConfig, err := dbConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := log.Default()

	userRepository := repositories.NewUserRepository(dbConfig)
	productRepository := repositories.NewProductRepository(dbConfig)
	subscriptionRespository := repositories.NewSubscriptionRepository(dbConfig)

	userService := app.NewUserService(userRepository)
	productService := app.NewProductService(productRepository)
	subscriptionService := app.NewSubscriptionService(subscriptionRespository, userRepository, productRepository)

	userHandler := handlers.NewUserHandler(logger, userService)
	productHandler := handlers.NewProductHandler(logger, productService)
	subscriptionHandler := handlers.NewSubscriptionHandler(logger, subscriptionService)

	engine := gin.Default()

	engine.POST("/users", userHandler.Create)
	engine.POST("/products", productHandler.Create)
	engine.GET("/products/:product-id", productHandler.Fetch)
	engine.GET("/products", productHandler.List)
	engine.POST("/subscriptions", subscriptionHandler.Create)
	engine.GET("/subscriptions/:subscription-id", subscriptionHandler.Fetch)
	engine.GET("/users/:user-id/subscriptions", subscriptionHandler.List)
	engine.PATCH("/subscriptions/:subscription-id", subscriptionHandler.Action)
	engine.Run()

	// router := mux.NewRouter()
	// router.PathPrefix("/swagger").Handler(http.FileServer(http.FS(swagger))).Methods(http.MethodGet)

	// log.Println("Starting server on port :8080")
	// http.ListenAndServe(":8080", router)
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
		domain.SubscriptionPlan{},
		domain.Product{},
		domain.Subscription{},
	)
	if err != nil {
		return nil, fmt.Errorf("could not migrate models: %w", err)
	}

	return db, err
}
