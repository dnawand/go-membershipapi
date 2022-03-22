package handlers

import (
	"errors"
	"net/http"

	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	logger *zap.Logger
	ss     domain.SubscriptionService
}

type subscribeRequest struct {
	UserID        string `json:"userId" binding:"required"`
	ProductID     string `json:"productId" binding:"required"`
	ProductPlanID string `json:"planId" binding:"required"`
}

type action string

const (
	Pause       action = "pause"
	Resume      action = "resume"
	Unsubscribe action = "unsubscribe"
)

type actionRequest struct {
	Action action `json:"action" binding:"required"`
}

func NewSubscriptionHandler(logger *zap.Logger, ss domain.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		logger: logger,
		ss:     ss,
	}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var request subscribeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("request binding error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	user, err := h.ss.Subscribe(request.UserID, request.ProductID, request.ProductPlanID, "")
	if err != nil {
		h.logger.Error("error when subscribing user to product", zap.Error(err), zap.Any("request", request))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *SubscriptionHandler) Fetch(c *gin.Context) {
	subscriptionID := c.Param("subscription-id")
	subscription, err := h.ss.Fetch(subscriptionID)
	if err != nil {
		var dataNotFoundError *domain.ErrDataNotFound

		if !errors.As(err, &dataNotFoundError) {
			h.logger.Debug("subscription not found", zap.Error(err), zap.String("subscriptionId", subscriptionID))
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		h.logger.Error("error when fetching subscription", zap.Error(err), zap.String("subscriptionId", subscriptionID))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	userID := c.Param("user-id")
	subscriptions, err := h.ss.List(userID)
	if err != nil {
		var dataNotFoundError *domain.ErrDataNotFound

		if !errors.As(err, &dataNotFoundError) {
			h.logger.Debug("subscriptions not found", zap.Error(err), zap.String("userId", userID))
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		h.logger.Error("error when listing subscriptions", zap.Error(err), zap.String("userId", userID))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) Action(c *gin.Context) {
	subscriptionID := c.Param("subscription-id")
	var request actionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("request binding error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	var subscription domain.Subscription
	var err error

	switch request.Action {
	case Pause:
		subscription, err = h.ss.Pause(subscriptionID)
	case Resume:
		subscription, err = h.ss.Resume(subscriptionID)
	case Unsubscribe:
		subscription, err = h.ss.Unsubscribe(subscriptionID)
	default:
		h.logger.Debug("invalid action on subscription", zap.Error(err), zap.Any("request", request))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	if err != nil {
		var errDataNotFound *domain.ErrDataNotFound

		if errors.As(err, &errDataNotFound) {
			h.logger.Debug("data not found", zap.Error(err), zap.Any("subscriptionId", subscriptionID))
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			h.logger.Debug("forbidden action", zap.Error(err), zap.Any("subscriptionId", subscriptionID), zap.Any("request", request))
			c.JSON(http.StatusForbidden, gin.H{})
			return
		}

		h.logger.Error(
			"error when performing action on subscription",
			zap.Error(err),
			zap.String("subscriptionId", subscriptionID),
			zap.Any("request", request),
		)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, subscription)
}
