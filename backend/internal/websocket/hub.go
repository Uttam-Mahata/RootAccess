package websocket

import (
	"encoding/json"
	"log"
	"sync"

	ws "github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	Hub    *Hub
	Conn   *ws.Conn
	Send   chan []byte
	UserID string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) BroadcastMessage(msgType string, payload interface{}) {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}
	h.broadcast <- data
}

func (h *Hub) SendToUser(userID string, msgType string, payload interface{}) {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling user message: %v", err)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(ws.CloseMessage, []byte{})
			return
		}
		if err := c.Conn.WriteMessage(ws.TextMessage, message); err != nil {
			return
		}
	}
}
