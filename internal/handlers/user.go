package handlers

import (
	"log"
	"net/http"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	logger      *log.Logger
	userService domain.UserService
}

func NewUserHandler(logger *log.Logger, userService domain.UserService) *UserHandler {
	return &UserHandler{
		logger:      logger,
		userService: userService,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var user domain.User

	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Println("error handling create user: %w", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.userService.Create(user)
	if err != nil {
		h.logger.Println("error when creating user: %w", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}
