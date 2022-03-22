package handlers

import (
	"net/http"

	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger      *zap.Logger
	userService domain.UserService
}

func NewUserHandler(logger *zap.Logger, userService domain.UserService) *UserHandler {
	return &UserHandler{
		logger:      logger,
		userService: userService,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var user domain.User

	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Error("error binding request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	user, err := h.userService.Create(user)
	if err != nil {
		h.logger.Error("error when creating user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusCreated, user)
}
