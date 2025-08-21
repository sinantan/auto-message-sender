package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sinan/auto-message-sender/internal/config"
	"github.com/sinan/auto-message-sender/internal/dataOperations"
	"github.com/sinan/auto-message-sender/internal/handlers"
	"github.com/sinan/auto-message-sender/pkg/logger"
	"github.com/sinan/auto-message-sender/pkg/mongodb"
	"github.com/sinan/auto-message-sender/pkg/redisdb"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	log := logger.New()
	log.Infof("Starting Auto Message Sender API (%s)", cfg.App.Environment)

	mongoClient := mongodb.New(cfg.MongoDB.URI, cfg.MongoDB.Database, cfg.App.Environment == "development")
	defer mongoClient.Close()

	redisConfig := &redisdb.RedisConfig{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		Database:     cfg.Redis.Database,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxConnAge:   cfg.Redis.MaxConnAge,
		PoolTimeout:  cfg.Redis.PoolTimeout,
		IdleTimeout:  cfg.Redis.IdleTimeout,
	}

	redisClient, err := redisdb.New(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	dataOps := dataOperations.New(mongoClient, redisClient, cfg)

	webhookHandler := handlers.NewWebhookHandler(cfg, log)
	messageHandler := handlers.NewMessageHandler(dataOps, cfg, log)
	schedulerHandler := handlers.NewSchedulerHandler(dataOps, cfg, log)
	schedulerHandler.SetWebhookHandler(webhookHandler)

	router := NewRouter(
		cfg,
		log,
		messageHandler,
		schedulerHandler,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Stop scheduler first and wait for jobs to complete
	if schedulerHandler != nil {
		log.Info("Stopping scheduler...")
		schedulerHandler.StopScheduler(&gin.Context{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited gracefully")
}
