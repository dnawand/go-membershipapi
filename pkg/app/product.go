package app

import (
	"github.com/dnawand/go-subscriptionapi/pkg/domain"
)

type ProductService struct {
	pr domain.ProductRepository
}

func NewProductService(pr domain.ProductRepository) *ProductService {
	return &ProductService{
		pr: pr,
	}
}

func (ps *ProductService) Create(product domain.Product) (domain.Product, error) {
	return ps.pr.Save(product)
}

func (ps *ProductService) Fetch(productID string) (domain.Product, error) {
	return ps.pr.Get(productID)
}

func (ps *ProductService) List() ([]domain.Product, error) {
	return ps.pr.List()
}
