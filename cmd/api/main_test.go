package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dnawand/go-membershipapi/internal/handlers"
	"github.com/dnawand/go-membershipapi/internal/storage"
	"github.com/dnawand/go-membershipapi/pkg/app"
	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/dnawand/go-membershipapi/pkg/repositories"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once sync.Once
var zapLogger *zap.Logger
var db *gorm.DB
var voucherStorage *storage.Store
var userRepository domain.UserRepository
var productRepository domain.ProductRepository
var subscriptionRespository domain.SubscriptionRepository

func TestListProducts(t *testing.T) {
	initContext()
	truncateTables()
	createdProducts := createProducts()

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		&handlers.SubscriptionHandler{},
	)
	req, _ := http.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var products []domain.Product
	json.Unmarshal(rr.Body.Bytes(), &products)

	assert.Equal(t, len(createdProducts), len(products))
}

func TestFetchSingleProduct(t *testing.T) {
	initContext()
	truncateTables()
	createdProducts := createProducts()
	expectedProduct := createdProducts[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		&handlers.SubscriptionHandler{},
	)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/products/%s", expectedProduct.ID), nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var product domain.Product
	json.Unmarshal(rr.Body.Bytes(), &product)

	assert.Equal(t, expectedProduct.ID, product.ID)
	assert.Equal(t, expectedProduct.Name, product.Name)
	assert.Equal(t, len(expectedProduct.ProductPlans), len(product.ProductPlans))
}

func TestSubscriptionNoVoucher(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	json.Unmarshal(rr.Body.Bytes(), &subscription)

	assert.NotEmpty(t, subscription.ID)
	assert.Equal(t, product.ID, subscription.Product.ID)
	assert.NotEqual(t, productPlan.ID, subscription.SubscriptionPlan.ID)
	assert.Equal(t, productPlan.Length, subscription.SubscriptionPlan.Length)
	assert.Equal(t, productPlan.Price.Code, subscription.SubscriptionPlan.Price.Code)
	assert.Equal(t, productPlan.Price.Number, subscription.SubscriptionPlan.Price.Number)
}

func TestFetchSubscription(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	json.Unmarshal(rr.Body.Bytes(), &subscription)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), nil)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	json.Unmarshal(rr.Body.Bytes(), &subscription)

	assert.NotEmpty(t, subscription.ID)
	assert.Equal(t, product.ID, subscription.Product.ID)
	assert.NotEqual(t, productPlan.ID, subscription.SubscriptionPlan.ID)
	assert.Equal(t, productPlan.Length, subscription.SubscriptionPlan.Length)
	assert.Equal(t, productPlan.Price.Code, subscription.SubscriptionPlan.Price.Code)
	assert.Equal(t, productPlan.Price.Number, subscription.SubscriptionPlan.Price.Number)
}

func TestSubscriptionPauseDuringTrial(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	json.Unmarshal(rr.Body.Bytes(), &subscription)

	jsonBody = `{"action": "pause"}`
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), strings.NewReader(jsonBody))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusLocked, rr.Code)
}

func repositoryAllowPauseOnTrial() domain.SubscriptionRepository {
	return &MockSubscriptionRepository{
		SaveFunc: func(u domain.User) (domain.Subscription, error) {
			return subscriptionRespository.Save(u)
		},
		GetFunc: func(subscriptionID string) (domain.Subscription, error) {
			s, _ := subscriptionRespository.Get(subscriptionID)
			// set TrialDate has it had already passed
			hours := 24 * 30 * app.TrialPeriod * 2
			s.TrialDate = s.TrialDate.Add(-time.Duration(hours) * time.Hour)
			return s, nil
		},
		UpdateFunc: func(s domain.Subscription, tu domain.ToUpdate) (domain.Subscription, error) {
			return subscriptionRespository.Update(s, tu)
		},
	}
}

func TestSubscriptionPauseAfterTrial(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			repositoryAllowPauseOnTrial(),
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	err := json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)

	assert.False(t, subscription.IsPaused)

	jsonBody = `{"action": "pause"}`
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), strings.NewReader(jsonBody))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.True(t, subscription.IsPaused)
}

func TestSubscriptionResumeAfterPause(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			repositoryAllowPauseOnTrial(),
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	err := json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.False(t, subscription.IsPaused)

	jsonBody = `{"action": "pause"}`
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), strings.NewReader(jsonBody))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.True(t, subscription.IsPaused)

	jsonBody = `{"action": "resume"}`
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), strings.NewReader(jsonBody))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.False(t, subscription.IsPaused)
}

func TestSubscriptionUnsubcribeOutOfTrial(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			repositoryAllowPauseOnTrial(),
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription

	err := json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.False(t, subscription.IsPaused)

	jsonBody = `{"action": "unsubscribe"}`
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), strings.NewReader(jsonBody))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.False(t, subscription.IsActive)
}

func TestSubscriptionUnsubcribeDuringTrial(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	jsonBody := fmt.Sprintf(`{"productId": "%s","planId": "%s"}`, product.ID, productPlan.ID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription

	err := json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.False(t, subscription.IsPaused)

	jsonBody = `{"action": "unsubscribe"}`
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%s/subscriptions/%s", user.ID, subscription.ID), strings.NewReader(jsonBody))
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &subscription)
	assert.NoError(t, err)
	assert.False(t, subscription.IsActive)
}

func TestSubscriptionCreationFixedAmountVoucher(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	fixedAmountVoucherID := "b86b4903-2043-4f71-b154-efec19fbc55a" // 5.00
	jsonBody := fmt.Sprintf(`{"productId": "%s", "planId": "%s", "voucherId": "%s"}`, product.ID, productPlan.ID, fixedAmountVoucherID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	json.Unmarshal(rr.Body.Bytes(), &subscription)

	assert.NotEmpty(t, subscription.ID)
	assert.Equal(t, product.ID, subscription.Product.ID)
	assert.NotEqual(t, productPlan.ID, subscription.SubscriptionPlan.ID)
	assert.Equal(t, productPlan.Length, subscription.SubscriptionPlan.Length)
	assert.Equal(t, "95.00", subscription.SubscriptionPlan.Price.Number) // fixed value discount: 5
	assert.Equal(t, productPlan.Price.Code, subscription.SubscriptionPlan.Price.Code)
	assert.Equal(t, "9.50", subscription.SubscriptionPlan.Tax.Number) // percentage discount: 5% = 0.50
	assert.Equal(t, productPlan.Tax.Code, subscription.SubscriptionPlan.Tax.Code)
}

func TestSubscriptionCreationPercentageVoucher(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	fixedAmountVoucherID := "4976ff21-a188-4bcc-97a0-2cf2278e9a6b" // 10.10
	jsonBody := fmt.Sprintf(`{"productId": "%s", "planId": "%s", "voucherId": "%s"}`, product.ID, productPlan.ID, fixedAmountVoucherID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var subscription domain.Subscription
	json.Unmarshal(rr.Body.Bytes(), &subscription)

	assert.NotEmpty(t, subscription.ID)
	assert.Equal(t, product.ID, subscription.Product.ID)
	assert.NotEqual(t, productPlan.ID, subscription.SubscriptionPlan.ID)
	assert.Equal(t, productPlan.Length, subscription.SubscriptionPlan.Length)
	assert.Equal(t, "89.90", subscription.SubscriptionPlan.Price.Number) // percentage discount: 10.10% = 10.10
	assert.Equal(t, productPlan.Price.Code, subscription.SubscriptionPlan.Price.Code)
	assert.Equal(t, "8.99", subscription.SubscriptionPlan.Tax.Number) // percentage discount: 10.10% = 1.01
	assert.Equal(t, productPlan.Tax.Code, subscription.SubscriptionPlan.Tax.Code)
}

func TestSubscriptionCreationInactiveVoucher(t *testing.T) {
	initContext()
	truncateTables()
	user := createUser()
	createdProducts := createProducts()
	product := createdProducts[0]
	productPlan := product.ProductPlans[0]

	router := configRouter(
		&handlers.UserHandler{},
		handlers.NewProductHandler(zapLogger, app.NewProductService(productRepository)),
		handlers.NewSubscriptionHandler(zapLogger, app.NewSubscriptionService(
			subscriptionRespository,
			userRepository,
			productRepository,
			voucherStorage,
			&app.DiscountService{},
		)),
	)

	fixedAmountVoucherID := "18c4b4ea-6fce-4ee7-8d3b-a16047a8789e" // inactive
	jsonBody := fmt.Sprintf(`{"productId": "%s", "planId": "%s", "voucherId": "%s"}`, product.ID, productPlan.ID, fixedAmountVoucherID)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/subscriptions", user.ID), strings.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func initContext() {
	once.Do(func() {
		zapLogger, _ = zap.NewDevelopment()
		db, _ = dbConfig()
		voucherStorage = loadVouchers()
		userRepository = repositories.NewUserRepository(db)
		productRepository = repositories.NewProductRepository(db)
		subscriptionRespository = repositories.NewSubscriptionRepository(db, voucherStorage)
	})
}

func createUser() domain.User {
	u, _ := userRepository.Save(domain.User{
		Name:  "Tester",
		Email: "tester@email.com",
	})
	return u
}

func createProducts() []domain.Product {
	p, _ := productRepository.Save(domain.Product{
		Name: "Test1",
		ProductPlans: []domain.ProductPlan{
			{Plan: &domain.Plan{Length: 1, Price: domain.Money{Code: domain.CurrencyEUR, Number: "100.00"}, Tax: domain.Money{Code: domain.CurrencyEUR, Number: "10.00"}}},
			{Plan: &domain.Plan{Length: 2, Price: domain.Money{Code: domain.CurrencyEUR, Number: "50.00"}, Tax: domain.Money{Code: domain.CurrencyEUR, Number: "5.50"}}},
		},
	})

	p2, _ := productRepository.Save(domain.Product{
		Name: "Test2",
		ProductPlans: []domain.ProductPlan{
			{Plan: &domain.Plan{Length: 1, Price: domain.Money{Code: domain.CurrencyEUR, Number: "12.99"}, Tax: domain.Money{Code: domain.CurrencyEUR, Number: "5.99"}}},
			{Plan: &domain.Plan{Length: 2, Price: domain.Money{Code: domain.CurrencyEUR, Number: "21.99"}, Tax: domain.Money{Code: domain.CurrencyEUR, Number: "4.99"}}},
			{Plan: &domain.Plan{Length: 3, Price: domain.Money{Code: domain.CurrencyEUR, Number: "20.99"}, Tax: domain.Money{Code: domain.CurrencyEUR, Number: "3.99"}}},
		},
	})

	return append([]domain.Product{}, p, p2)
}

func truncateTables() {
	db.Exec("TRUNCATE TABLE subscriptions CASCADE;")
	db.Exec("TRUNCATE TABLE products CASCADE;")
	db.Exec("TRUNCATE TABLE users CASCADE;")
}

type MockSubscriptionRepository struct {
	SaveFunc   func(domain.User) (domain.Subscription, error)
	GetFunc    func(subscriptionID string) (domain.Subscription, error)
	ListFunc   func(userID string) ([]domain.Subscription, error)
	UpdateFunc func(domain.Subscription, domain.ToUpdate) (domain.Subscription, error)
}

func (msr *MockSubscriptionRepository) Save(u domain.User) (domain.Subscription, error) {
	return msr.SaveFunc(u)
}

func (msr *MockSubscriptionRepository) Get(subscriptionID string) (domain.Subscription, error) {
	return msr.GetFunc(subscriptionID)
}

func (msr *MockSubscriptionRepository) List(userID string) ([]domain.Subscription, error) {
	return msr.ListFunc(userID)
}

func (msr *MockSubscriptionRepository) Update(s domain.Subscription, toUpdate domain.ToUpdate) (domain.Subscription, error) {
	return msr.UpdateFunc(s, toUpdate)
}
