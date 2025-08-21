package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/internal/handlers"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	cfg *config.Config,
	logger *logrus.Logger,
	messageHandler *handlers.MessageHandler,
	schedulerHandler *handlers.SchedulerHandler,
	webhookHandler *handlers.WebhookHandler,
) *gin.Engine {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Auto Message Sender API",
			"environment": cfg.App.Environment,
			"version":     "1.0.0",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": gin.H{
				"unix": gin.H{
					"seconds": gin.H{
						"value": "now",
					},
				},
			},
		})
	})

	api := router.Group("/api/v1")
	{
		messages := api.Group("/messages")
		{
			messages.POST("", messageHandler.CreateMessage)
			messages.GET("/sent", messageHandler.GetSentMessages)
		}

		scheduler := api.Group("/scheduler")
		{
			scheduler.POST("/start", schedulerHandler.StartScheduler)
			scheduler.POST("/stop", schedulerHandler.StopScheduler)
			scheduler.GET("/status", schedulerHandler.GetSchedulerStatus)
		}
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func loggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := gin.H{
			"unix": gin.H{
				"seconds": gin.H{
					"value": "now",
				},
			},
		}

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := gin.H{
			"unix": gin.H{
				"seconds": gin.H{
					"value": "now",
				},
			},
		}

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"start_time":  start,
		}).Info("HTTP Request")
	}
}
