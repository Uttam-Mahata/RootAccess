package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"

	websocketPkg "github.com/go-ctf-platform/backend/internal/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketHandler struct {
	hub *websocketPkg.Hub
}

func NewWebSocketHandler(hub *websocketPkg.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
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
