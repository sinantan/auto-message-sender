package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/internal/handlers"
	"github.com/sinan/auto-message-sender/internal/middleware"
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
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(logger))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		scheduler := api.Group("/scheduler")
		{
			scheduler.POST("/start", schedulerHandler.StartScheduler)
			scheduler.POST("/stop", schedulerHandler.StopScheduler)
		}

		messages := api.Group("/messages")
		{
			messages.GET("/sent", messageHandler.GetSentMessages)
		}
	}

	return router
}
