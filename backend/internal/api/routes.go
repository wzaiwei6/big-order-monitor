package api

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/sessions", h.GetSessions)
		api.POST("/connect", h.Connect)
		api.POST("/disconnect", h.Disconnect)
	}

	r.GET("/ws", h.WebSocket)
}

