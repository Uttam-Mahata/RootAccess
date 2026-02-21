package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	websocketPkg "github.com/Uttam-Mahata/RootAccess/backend/internal/websocket"
)

type WebSocketHandler struct {
	hub      websocketPkg.Hub
	upgrader ws.Upgrader
}

func NewWebSocketHandler(hub websocketPkg.Hub, cfg *config.Config) *WebSocketHandler {
	allowedOrigin := cfg.FrontendURL
	return &WebSocketHandler{
		hub: hub,
		upgrader: ws.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == "" || origin == allowedOrigin
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket (for standard servers)
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)

	client := &websocketPkg.Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userIDStr,
	}

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// HandleLambdaConnect handles API Gateway WebSocket $connect events
func (h *WebSocketHandler) HandleLambdaConnect(c *gin.Context) {
	connID := c.GetHeader("X-Connection-Id")
	userID := c.GetHeader("X-User-Id")

	if connID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing connection id"})
		return
	}

	if err := h.hub.RegisterConnection(context.Background(), connID, userID); err != nil {
		log.Printf("Failed to register lambda connection %s: %v", connID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register connection"})
		return
	}

	c.Status(http.StatusOK)
}

// HandleLambdaDisconnect handles API Gateway WebSocket $disconnect events
func (h *WebSocketHandler) HandleLambdaDisconnect(c *gin.Context) {
	connID := c.GetHeader("X-Connection-Id")

	if connID != "" {
		h.hub.UnregisterConnection(context.Background(), connID)
	}

	c.Status(http.StatusOK)
}

// HandleLambdaDefault handles API Gateway WebSocket $default events
func (h *WebSocketHandler) HandleLambdaDefault(c *gin.Context) {
	c.Status(http.StatusOK)
}
