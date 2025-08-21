package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

// CreateMessage godoc
// @Summary Create a new message
// @Description Create a new message to be sent
// @Tags messages
// @Accept json
// @Produce json
// @Param message body models.CreateMessageRequest true "Message data"
// @Success 201 {object} models.Message
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req models.CreateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithError(err).Error("Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	message := models.NewMessage(req.To, req.Content)
	message.ID = uuid.New().String()

	if err := h.dataOps.CreateMessage(message); err != nil {
		h.logger.WithError(err).Error("Failed to create message")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create message",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"to":         message.To,
	}).Info("Message created successfully")

	c.JSON(http.StatusCreated, message)
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
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
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
