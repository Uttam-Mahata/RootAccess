package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	ws "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// UserMessage represents a message for a specific user across instances
type UserMessage struct {
	UserID  string      `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	Hub    Hub
	Conn   *ws.Conn
	Send   chan []byte
	UserID string
}

type Hub interface {
	Run()
	Register(client *Client)
	BroadcastMessage(msgType string, payload interface{})
	SendToUser(userID string, msgType string, payload interface{})
	// AWS Lambda specific methods
	RegisterConnection(ctx context.Context, connectionID string, userID string) error
	UnregisterConnection(ctx context.Context, connectionID string) error
}

// MemoryHub: Standard implementation for persistent servers (EC2/Local)
type MemoryHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() Hub {
	return &MemoryHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *MemoryHub) Run() {
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

func (h *MemoryHub) Register(client *Client) {
	h.register <- client
}

func (h *MemoryHub) BroadcastMessage(msgType string, payload interface{}) {
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

func (h *MemoryHub) SendToUser(userID string, msgType string, payload interface{}) {
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

// Stub AWS methods for MemoryHub
func (h *MemoryHub) RegisterConnection(ctx context.Context, c string, u string) error { return nil }
func (h *MemoryHub) UnregisterConnection(ctx context.Context, c string) error        { return nil }

// RedisHub: Distributed implementation for persistent servers (EC2/Containers)
type RedisHub struct {
	*MemoryHub
	redisClient *redis.Client
	broadcastChannel string
	userChannel      string
}

func NewRedisHub(redisClient *redis.Client) Hub {
	return &RedisHub{
		MemoryHub:        NewHub().(*MemoryHub),
		redisClient:      redisClient,
		broadcastChannel: "ws_broadcast",
		userChannel:      "ws_user",
	}
}

func (h *RedisHub) Run() {
	go h.MemoryHub.Run()
	ctx := context.Background()
	pubsub := h.redisClient.Subscribe(ctx, h.broadcastChannel, h.userChannel)
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		if msg.Channel == h.broadcastChannel {
			h.MemoryHub.broadcast <- []byte(msg.Payload)
		} else if msg.Channel == h.userChannel {
			var userMsg UserMessage
			if err := json.Unmarshal([]byte(msg.Payload), &userMsg); err == nil {
				h.MemoryHub.SendToUser(userMsg.UserID, userMsg.Type, userMsg.Payload)
			}
		}
	}
}

func (h *RedisHub) BroadcastMessage(msgType string, payload interface{}) {
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.redisClient.Publish(context.Background(), h.broadcastChannel, data)
}

func (h *RedisHub) SendToUser(userID string, msgType string, payload interface{}) {
	userMsg := UserMessage{
		UserID:  userID,
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(userMsg)
	if err != nil {
		return
	}
	h.redisClient.Publish(context.Background(), h.userChannel, data)
}

// AwsLambdaHub: Stateless implementation using API Gateway WebSocket API + Redis
type AwsLambdaHub struct {
	redisClient *redis.Client
	awsClient   *apigatewaymanagementapi.Client
	callbackURL string
	connectionsSetKey string // Redis set key for all connection IDs
	userPrefixKey      string // Redis key prefix for user-to-connections mapping
}

func NewAwsLambdaHub(redisClient *redis.Client, awsCfg aws.Config, callbackURL string) Hub {
	// API Gateway Management API client
	apiClient := apigatewaymanagementapi.NewFromConfig(awsCfg, func(o *apigatewaymanagementapi.Options) {
		if callbackURL != "" {
			o.BaseEndpoint = aws.String(callbackURL)
		}
	})

	return &AwsLambdaHub{
		redisClient:       redisClient,
		awsClient:         apiClient,
		callbackURL:       callbackURL,
		connectionsSetKey: "ws:active_connections",
		userPrefixKey:      "ws:user_connections:",
	}
}

func (h *AwsLambdaHub) Run() { /* Stateless - no loop needed */ }
func (h *AwsLambdaHub) Register(client *Client) { /* Handled by API Gateway $connect */ }

func (h *AwsLambdaHub) RegisterConnection(ctx context.Context, connectionID string, userID string) error {
	// 1. Add to global active connections set
	if err := h.redisClient.SAdd(ctx, h.connectionsSetKey, connectionID).Err(); err != nil {
		return err
	}
	// 2. If userID provided, add to user-specific set
	if userID != "" {
		userKey := h.userPrefixKey + userID
		if err := h.redisClient.SAdd(ctx, userKey, connectionID).Err(); err != nil {
			return err
		}
		// Set expiry for user mapping (e.g., 24h)
		h.redisClient.Expire(ctx, userKey, 24*time.Hour)
	}
	return nil
}

func (h *AwsLambdaHub) UnregisterConnection(ctx context.Context, connectionID string) error {
	// Remove from global set
	h.redisClient.SRem(ctx, h.connectionsSetKey, connectionID)
	// Note: Removing from user-specific sets is harder without storing the mapping.
	// We'll rely on the global check during broadcast to clean up.
	return nil
}

func (h *AwsLambdaHub) BroadcastMessage(msgType string, payload interface{}) {
	ctx := context.Background()
	msg := Message{Type: msgType, Payload: payload}
	data, _ := json.Marshal(msg)

	// Get all active connection IDs from Redis
	connIDs, err := h.redisClient.SMembers(ctx, h.connectionsSetKey).Result()
	if err != nil {
		log.Printf("Error fetching connection IDs: %v", err)
		return
	}

	for _, id := range connIDs {
		h.sendToConnection(ctx, id, data)
	}
}

func (h *AwsLambdaHub) SendToUser(userID string, msgType string, payload interface{}) {
	ctx := context.Background()
	msg := Message{Type: msgType, Payload: payload}
	data, _ := json.Marshal(msg)

	userKey := h.userPrefixKey + userID
	connIDs, _ := h.redisClient.SMembers(ctx, userKey).Result()

	for _, id := range connIDs {
		h.sendToConnection(ctx, id, data)
	}
}

func (h *AwsLambdaHub) sendToConnection(ctx context.Context, connectionID string, data []byte) {
	_, err := h.awsClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         data,
	})

	if err != nil {
		// If connection is gone, remove it from Redis
		log.Printf("Error sending to %s: %v", connectionID, err)
		h.UnregisterConnection(ctx, connectionID)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		if mh, ok := c.Hub.(*MemoryHub); ok {
			mh.unregister <- c
		}
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
