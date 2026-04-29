package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ordermonitor/internal/config"
	"ordermonitor/internal/summary"
	ws "ordermonitor/internal/websocket"
	"ordermonitor/pkg/logger"
)

type Handler struct {
	cfg            config.Config
	log            *zap.Logger
	wsManager      *ws.Manager
	summaryManager *summary.Manager
	redis          *redis.Client
	db             *sql.DB
}

func NewHandler(cfg config.Config, log *zap.Logger, wsManager *ws.Manager, summaryManager *summary.Manager, redis *redis.Client, db *sql.DB) *Handler {
	return &Handler{cfg: cfg, log: log, wsManager: wsManager, summaryManager: summaryManager, redis: redis, db: db}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetSessions(c *gin.Context) {
	sessions := h.wsManager.ActiveSessions()
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *Handler) Connect(c *gin.Context) {
	var req ws.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid connect payload", logger.ErrorField(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.wsManager.StartSession(c.Request.Context(), req); err != nil {
		h.log.Error("failed to start session", logger.ErrorField(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session started"})
}

func (h *Handler) Disconnect(c *gin.Context) {
	type payload struct {
		Key string `json:"key" binding:"required"`
	}

	var req payload
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid disconnect payload", logger.ErrorField(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.wsManager.StopSession(req.Key)
	c.JSON(http.StatusOK, gin.H{"message": "session stopped"})
}

func (h *Handler) StartSummary(c *gin.Context) {
	if err := h.summaryManager.Start(); err != nil {
		h.log.Error("summary start failed", logger.ErrorField(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "summary started"})
}

func (h *Handler) StopSummary(c *gin.Context) {
	h.summaryManager.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "summary stopped"})
}

