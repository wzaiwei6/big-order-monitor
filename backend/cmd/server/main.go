package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ordermonitor/internal/api"
	"ordermonitor/internal/config"
	"ordermonitor/internal/storage"
	"ordermonitor/internal/summary"
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

	db, err := storage.NewSQLite(cfg.SQLite)
	if err != nil {
		logr.Fatal("failed to init sqlite", logger.ErrorField(err))
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

	summaryManager := summary.NewManager(cfg, logr, db)
	if err := summaryManager.Start(); err != nil {
		logr.Fatal("failed to start collector", logger.ErrorField(err))
	}

	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()
	go runRetentionCleanup(runtimeCtx, cfg, db, logr)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.LoggingMiddleware(logr))

	handler := api.NewHandler(cfg, logr, wsManager, summaryManager, redisClient, db)
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

	runtimeCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logr.Info("server shutting down")
	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatal("server forced to shutdown", logger.ErrorField(err))
	}

	wsManager.Shutdown()
	summaryManager.Stop()

	logr.Info("goodbye")
}

func runRetentionCleanup(ctx context.Context, cfg config.Config, db *sql.DB, logr *zap.Logger) {
	interval := time.Duration(cfg.Monitor.CleanupIntervalMinute) * time.Minute
	if interval <= 0 {
		interval = time.Hour
	}

	retention := time.Duration(cfg.Monitor.DataRetentionHours) * time.Hour
	if retention <= 0 {
		retention = 12 * time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	lastVacuumDate := ""

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			cutoff := now.Add(-retention).Unix()
			rows, err := storage.CleanupOrdersBefore(db, cutoff)
			if err != nil {
				logr.Warn("cleanup expired orders failed", logger.ErrorField(err))
			} else if rows > 0 {
				logr.Info("cleanup expired orders finished", zap.Int64("deleted", rows))
			}

			if shouldRunVacuum(now, cfg.SQLite, lastVacuumDate) {
				if err := storage.VacuumSQLite(db); err != nil {
					logr.Warn("sqlite vacuum failed", logger.ErrorField(err))
				} else {
					lastVacuumDate = now.Format("2006-01-02")
					logr.Info("sqlite vacuum finished",
						zap.String("date", lastVacuumDate),
						zap.Int("retention_hours", cfg.Monitor.DataRetentionHours),
					)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func shouldRunVacuum(now time.Time, cfg config.SQLiteConfig, lastVacuumDate string) bool {
	if !cfg.VacuumEnabled {
		return false
	}

	today := now.Format("2006-01-02")
	if today == lastVacuumDate {
		return false
	}

	hour := cfg.VacuumHour
	if hour < 0 || hour > 23 {
		hour = 4
	}

	minute := cfg.VacuumMinute
	if minute < 0 || minute > 59 {
		minute = 0
	}

	if now.Hour() < hour {
		return false
	}

	if now.Hour() == hour && now.Minute() < minute {
		return false
	}

	return true
}
