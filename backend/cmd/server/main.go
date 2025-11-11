package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"ordermonitor/internal/api"
	"ordermonitor/internal/config"
	"ordermonitor/internal/storage"
	"ordermonitor/internal/websocket"
	"ordermonitor/pkg/logger"
)

func main() {
	cfg := config.Load()

	logr, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logr.Sync() //nolint:errcheck

	db, err := storage.NewMySQL(cfg.MySQL)
	if err != nil {
		logr.Fatal("failed to init mysql", logger.ErrorField(err))
	}
	defer db.Close()

	redisClient, err := storage.NewRedis(cfg.Redis)
	if err != nil {
		logr.Fatal("failed to init redis", logger.ErrorField(err))
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	wsManager := websocket.NewManager(cfg, logr, redisClient, db)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.LoggingMiddleware(logr))

	handler := api.NewHandler(cfg, logr, wsManager, redisClient, db)
	api.RegisterRoutes(r, handler)

	srv := &http.Server{
		Addr:           cfg.Server.ListenAddr(),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logr.Info("server starting", logger.String("addr", cfg.Server.ListenAddr()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Fatal("server exited", logger.ErrorField(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logr.Info("server shutting down")
	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatal("server forced to shutdown", logger.ErrorField(err))
	}

	wsManager.Shutdown()

	logr.Info("goodbye")
}

