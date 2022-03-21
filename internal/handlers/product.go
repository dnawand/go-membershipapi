package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	logger *log.Logger
	ps     domain.ProductService
}

func NewProductHandler(logger *log.Logger, ps domain.ProductService) *ProductHandler {
	return &ProductHandler{
		logger: logger,
		ps:     ps,
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var product domain.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		h.logger.Printf("error handling create product: %s\n", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.ps.Create(product)
	if err != nil {
		h.logger.Printf("error when creating product: %s\n", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *ProductHandler) Fetch(c *gin.Context) {
	productID := c.Param("product-id")

	product, err := h.ps.Fetch(productID)
	if err != nil {
		var dataNotFoundError *domain.DataNotFoundError

		if !errors.As(err, &dataNotFoundError) {
			h.logger.Printf("error when fetching product: %s\n", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.ps.List()
	if err != nil {
		var dataNotFoundError *domain.DataNotFoundError

		if !errors.As(err, &dataNotFoundError) {
			h.logger.Printf("error when listing products: %s\n", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, products)
}
