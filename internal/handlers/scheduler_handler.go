package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/internal/dataOperations"
	"github.com/sinan/auto-message-sender/internal/models"
	"github.com/sirupsen/logrus"
)

type SchedulerHandler struct {
	dataOps        *dataOperations.DataOperations
	config         *config.Config
	logger         *logrus.Logger
	webhookHandler *WebhookHandler
	ticker         *time.Ticker
	stopChan       chan struct{}
	isRunning      bool
	activeJobs     sync.WaitGroup
	mu             sync.RWMutex
}

func NewSchedulerHandler(dataOps *dataOperations.DataOperations, config *config.Config, logger *logrus.Logger) *SchedulerHandler {
	return &SchedulerHandler{
		dataOps:   dataOps,
		config:    config,
		logger:    logger,
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

func (h *SchedulerHandler) SetWebhookHandler(webhookHandler *WebhookHandler) {
	h.webhookHandler = webhookHandler
}

// StartScheduler godoc
// @Summary Start the message scheduler
// @Description Start the automatic message sending scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} models.SchedulerResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /scheduler/start [post]
func (h *SchedulerHandler) StartScheduler(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.isRunning {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Scheduler is already running",
		})
		return
	}

	if err := h.dataOps.UpdateSchedulerStatus(true); err != nil {
		h.logger.WithError(err).Error("Failed to update scheduler status")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start scheduler",
		})
		return
	}

	h.startScheduler()

	h.logger.Info("Scheduler started successfully")

	now := time.Now()
	response := models.SchedulerResponse{
		IsActive:  true,
		StartedAt: &now,
		Message:   "Scheduler started successfully",
	}

	c.JSON(http.StatusOK, response)
}

// StopScheduler godoc
// @Summary Stop the message scheduler
// @Description Stop the automatic message sending scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} models.SchedulerResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /scheduler/stop [post]
func (h *SchedulerHandler) StopScheduler(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isRunning {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Scheduler is not running",
		})
		return
	}

	if err := h.dataOps.UpdateSchedulerStatus(false); err != nil {
		h.logger.WithError(err).Error("Failed to update scheduler status")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop scheduler",
		})
		return
	}

	h.stopScheduler()

	h.logger.Info("Scheduler stopped successfully")

	now := time.Now()
	response := models.SchedulerResponse{
		IsActive:  false,
		StoppedAt: &now,
		Message:   "Scheduler stopped successfully",
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchedulerHandler) startScheduler() {
	h.isRunning = true
	h.stopChan = make(chan struct{})
	h.ticker = time.NewTicker(h.config.App.SchedulerInterval)

	go func() {
		h.logger.WithField("interval", h.config.App.SchedulerInterval).Info("Message scheduler started")

		for {
			select {
			case <-h.ticker.C:
				h.processMessages()
			case <-h.stopChan:
				h.logger.Info("Scheduler stopped")
				return
			}
		}
	}()
}

func (h *SchedulerHandler) stopScheduler() {
	if h.ticker != nil {
		h.ticker.Stop()
	}
	close(h.stopChan)

	h.logger.Info("Waiting for active jobs to complete...")
	h.activeJobs.Wait()
	h.logger.Info("All jobs completed, scheduler stopped")

	h.isRunning = false
}

func (h *SchedulerHandler) processMessages() {
	ctx := context.Background()

	messages, err := h.dataOps.GetPendingMessages(h.config.App.MessagesPerInterval)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pending messages")
		return
	}

	if len(messages) == 0 {
		h.logger.Debug("No pending messages to process")
		return
	}

	h.logger.WithField("message_count", len(messages)).Info("Processing messages")

	for _, message := range messages {
		h.activeJobs.Add(1)
		go func(msg models.Message) {
			defer h.activeJobs.Done()
			h.processSingleMessage(ctx, msg)
		}(message)
	}
}

func (h *SchedulerHandler) processSingleMessage(ctx context.Context, message models.Message) {
	logger := h.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"to":         message.To,
	})

	logger.Info("Sending message")

	webhookReq := models.WebhookRequest{
		To:      message.To,
		Content: message.Content,
	}

	// Try with retry mechanism
	response, err := h.webhookHandler.SendMessageWithRetry(ctx, webhookReq, 3)
	if err != nil {
		logger.WithError(err).Error("Failed to send message after retries")

		errorMsg := err.Error()
		if err := h.dataOps.UpdateMessageStatus(message.ID, models.MessageStatusFailed, nil, &errorMsg); err != nil {
			logger.WithError(err).Error("Failed to update message status to failed")
		}
		return
	}

	logger.WithField("webhook_message_id", response.MessageID).Info("Message sent successfully")

	if err := h.dataOps.UpdateMessageStatus(message.ID, models.MessageStatusSent, &response.MessageID, nil); err != nil {
		logger.WithError(err).Error("Failed to update message status to sent")
		return
	}

	if err := h.dataOps.CacheMessage(response.MessageID, time.Now()); err != nil {
		logger.WithError(err).Warn("Failed to cache message (non-critical)")
	}
}
