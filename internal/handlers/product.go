package handlers

import (
	"errors"
	"net/http"

	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	logger *zap.Logger
	ps     domain.ProductService
}

func NewProductHandler(logger *zap.Logger, ps domain.ProductService) *ProductHandler {
	return &ProductHandler{
		logger: logger,
		ps:     ps,
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var product domain.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		h.logger.Error("request binding error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	user, err := h.ps.Create(product)
	if err != nil {
		h.logger.Error("error when creating product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *ProductHandler) Fetch(c *gin.Context) {
	productID := c.Param("product-id")

	product, err := h.ps.Fetch(productID)
	if err != nil {
		var dataNotFoundError *domain.ErrDataNotFound

		if !errors.As(err, &dataNotFoundError) {
			h.logger.Debug("product not found", zap.Error(err), zap.String("productId", productID))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		h.logger.Error("error when fetching product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.ps.List()
	if err != nil {
		var dataNotFoundError *domain.ErrDataNotFound

		if !errors.As(err, &dataNotFoundError) {
			h.logger.Debug("products not found", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		h.logger.Error("error when listing products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, products)
}
