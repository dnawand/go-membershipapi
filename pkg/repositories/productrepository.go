package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (pr *ProductRepository) Save(product domain.Product) (domain.Product, error) {
	now := time.Now()
	productID, err := uuid.NewRandom()
	if err != nil {
		return domain.Product{}, fmt.Errorf("error when generating id for user: %w", err)
	}
	product.ID = productID.String()
	product.CreatedAt = now
	product.UpdatedAt = now

	for i := range product.SubscriptionPlans {
		id, err := uuid.NewRandom()
		if err != nil {
			return domain.Product{}, fmt.Errorf("error when generating id for subscription plan: %w", err)
		}
		product.SubscriptionPlans[i].ID = id.String()
		product.SubscriptionPlans[i].CreatedAt = now
		product.SubscriptionPlans[i].UpdatedAt = now
		product.SubscriptionPlans[i].ProductID = product.ID
	}

	if tx := pr.db.Create(product); tx.Error != nil {
		return domain.Product{}, fmt.Errorf("could not save new product: %w", tx.Error)
	}

	return product, nil
}

func (pr *ProductRepository) Get(productID string) (domain.Product, error) {
	var product domain.Product

	if tx := pr.db.Preload("SubscriptionPlans").Find(&product, "id = ?", productID); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return domain.Product{}, &domain.DataNotFoundError{DataType: "product"}
		}
		return domain.Product{}, fmt.Errorf("error when querying product: %w", tx.Error)
	}

	return product, nil
}

func (pr *ProductRepository) List() ([]domain.Product, error) {
	var products = []domain.Product{}

	if tx := pr.db.Preload("SubscriptionPlans").Find(&products); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) || len(products) == 0 {
			return products, &domain.DataNotFoundError{DataType: "product list"}
		}
		return nil, fmt.Errorf("error when querying products: %w", tx.Error)
	}

	return products, nil
}
