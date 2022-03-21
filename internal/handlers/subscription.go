package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	logger *log.Logger
	ss     domain.SubscriptionService
}

type subscribeRequest struct {
	UserID             string `json:"userId" binding:"required"`
	ProductID          string `json:"productId" binding:"required"`
	SubscriptionPlanID string `json:"subscriptionPlanId" binding:"required"`
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

func NewSubscriptionHandler(logger *log.Logger, ss domain.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		logger: logger,
		ss:     ss,
	}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var request subscribeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("error handling create subscription: %s\n", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.ss.Subscribe(request.UserID, request.ProductID, request.SubscriptionPlanID)
	if err != nil {
		h.logger.Printf("error when subscribing user to product: %s\n", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			h.logger.Printf("error when fetching product: %s\n", err.Error())

			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
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
			h.logger.Printf("error when listing subscritions: %s\n", err.Error())

			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) Action(c *gin.Context) {
	subscriptionID := c.Param("subscription-id")
	var request actionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("error handling subscription command: %s\n", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	if err != nil {
		var errDataNotFound *domain.ErrDataNotFound

		if !errors.As(err, &errDataNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		if !errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		h.logger.Printf("error when action (%s) on subscription: %s\n", request.Action, err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}
