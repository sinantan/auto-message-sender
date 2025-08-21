package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/internal/models"
	"github.com/sirupsen/logrus"
)

type WebhookHandler struct {
	config     *config.Config
	logger     *logrus.Logger
	httpClient *http.Client
}

func NewWebhookHandler(config *config.Config, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: config.Webhook.Timeout,
		},
	}
}

func (h *WebhookHandler) SendMessage(ctx context.Context, request models.WebhookRequest) (*models.WebhookResponse, error) {
	logger := h.logger.WithFields(logrus.Fields{
		"to":      request.To,
		"content": request.Content,
		"url":     h.config.Webhook.URL,
	})

	jsonData, err := json.Marshal(request)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.config.Webhook.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.WithError(err).Error("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ins-auth-key", h.config.Webhook.AuthKey)

	startTime := time.Now()
	resp, err := h.httpClient.Do(req)
	duration := time.Since(startTime)

	logger = logger.WithFields(logrus.Fields{
		"duration_ms": duration.Milliseconds(),
		"status_code": func() int {
			if resp != nil {
				return resp.StatusCode
			}
			return 0
		}(),
	})

	if err != nil {
		logger.WithError(err).Error("HTTP request failed")
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Error("Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		logger.WithFields(logrus.Fields{
			"response_body": string(body),
		}).Error("Webhook returned non-success status")
		return nil, fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	var webhookResponse models.WebhookResponse
	if err := json.Unmarshal(body, &webhookResponse); err != nil {
		logger.WithError(err).WithField("response_body", string(body)).Error("Failed to unmarshal response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.WithField("message_id", webhookResponse.MessageID).Info("Webhook request successful")

	return &webhookResponse, nil
}

func (h *WebhookHandler) SendMessageWithRetry(ctx context.Context, request models.WebhookRequest, maxRetries int) (*models.WebhookResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger := h.logger.WithFields(logrus.Fields{
			"attempt":     attempt,
			"max_retries": maxRetries,
			"to":          request.To,
		})

		response, err := h.SendMessage(ctx, request)
		if err == nil {
			if attempt > 1 {
				logger.Info("Message sent successfully after retry")
			}
			return response, nil
		}

		lastErr = err
		logger.WithError(err).Warn("Message send attempt failed")

		if attempt < maxRetries {
			backoffDuration := time.Duration(attempt) * 5 * time.Second
			logger.WithField("backoff_seconds", backoffDuration.Seconds()).Info("Retrying after backoff")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
			}
		}
	}

	h.logger.WithError(lastErr).WithFields(logrus.Fields{
		"attempts": maxRetries,
		"to":       request.To,
	}).Error("All retry attempts failed")

	return nil, fmt.Errorf("all %d attempts failed, last error: %w", maxRetries, lastErr)
}
