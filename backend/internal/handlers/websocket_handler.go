package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"

	"github.com/go-ctf-platform/backend/internal/config"
	websocketPkg "github.com/go-ctf-platform/backend/internal/websocket"
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
