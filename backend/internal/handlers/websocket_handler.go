package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	websocketPkg "github.com/Uttam-Mahata/RootAccess/backend/internal/websocket"
)

type WebSocketHandler struct {
	hub      *websocketPkg.Hub
	upgrader ws.Upgrader
}

func NewWebSocketHandler(hub *websocketPkg.Hub, cfg *config.Config) *WebSocketHandler {
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

// HandleWebSocket upgrades HTTP connection to WebSocket
// @Summary WebSocket connection
// @Description Establish a WebSocket connection for real-time updates (solves, scoreboard updates).
// @Tags WebSocket
// @Success 101 {string} string "Switching Protocols"
// @Router /ws [get]
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
