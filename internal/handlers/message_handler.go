package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/internal/dataOperations"
	"github.com/sinan/auto-message-sender/internal/models"
	"github.com/sirupsen/logrus"
)

type MessageHandler struct {
	dataOps   *dataOperations.DataOperations
	config    *config.Config
	logger    *logrus.Logger
	validator *validator.Validate
}

func NewMessageHandler(dataOps *dataOperations.DataOperations, config *config.Config, logger *logrus.Logger) *MessageHandler {
	return &MessageHandler{
		dataOps:   dataOps,
		config:    config,
		logger:    logger,
		validator: validator.New(),
	}
}

// GetSentMessages godoc
// @Summary Get list of sent messages
// @Description Retrieve a paginated list of sent messages
// @Tags messages
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} models.MessageListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/messages/sent [get]
func (h *MessageHandler) GetSentMessages(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		h.logger.WithError(err).Warn("Invalid page parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page parameter, must be a positive integer",
		})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if err != nil || perPage < 1 {
		h.logger.WithError(err).Warn("Invalid per_page parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid per_page parameter, must be a positive integer",
		})
		return
	}

	if perPage > 100 {
		h.logger.Warn("per_page parameter too large, limiting to 100")
		perPage = 100
	}

	messages, total, err := h.dataOps.GetSentMessages(page, perPage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get sent messages")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve messages",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	response := models.MessageListResponse{
		Messages:   messages,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	h.logger.WithFields(logrus.Fields{
		"total_messages": total,
		"page":           page,
		"per_page":       perPage,
	}).Info("Retrieved sent messages")

	c.JSON(http.StatusOK, response)
}
