package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// RedisRegistry holds purpose-specific Redis clients for distributed caching.
// Each instance maps to a separate Upstash Redis database to stay within free-tier limits.
type RedisRegistry struct {
	Auth       *redis.Client // OAuth state, session tokens
	Scoreboard *redis.Client // Scoreboard + team scoreboard cache
	Challenge  *redis.Client // Challenge-related caching (solve tracking, attempts)
	WebSocket  *redis.Client // WebSocket connection tracking (Lambda hub) or pub/sub (Redis hub)
	Analytics  *redis.Client // Platform analytics cache
	General    *redis.Client // Misc/overflow, config metadata
}

// Registry is the global multi-instance Redis registry.
// Nil if Redis is not configured.
var Registry *RedisRegistry

// ConnectRedisRegistry initializes all 6 Redis instances from URL connection strings.
// URLs should be in the format: rediss://default:token@host:6379
// If a URL is empty, that instance is set to nil (graceful degradation).
// If all URLs are empty, the entire Registry is set to nil.
func ConnectRedisRegistry(urls map[string]string) {
	reg := &RedisRegistry{}
	connected := 0

	reg.Auth = connectFromURL("Auth", urls["auth"])
	if reg.Auth != nil {
		connected++
	}

	reg.Scoreboard = connectFromURL("Scoreboard", urls["scoreboard"])
	if reg.Scoreboard != nil {
		connected++
	}

	reg.Challenge = connectFromURL("Challenge", urls["challenge"])
	if reg.Challenge != nil {
		connected++
	}

	reg.WebSocket = connectFromURL("WebSocket", urls["websocket"])
	if reg.WebSocket != nil {
		connected++
	}

	reg.Analytics = connectFromURL("Analytics", urls["analytics"])
	if reg.Analytics != nil {
		connected++
	}

	reg.General = connectFromURL("General", urls["general"])
	if reg.General != nil {
		connected++
	}

	if connected == 0 {
		log.Println("Warning: No Redis instances connected. Registry disabled.")
		Registry = nil
		return
	}

	log.Printf("Redis Registry: %d/6 instances connected", connected)
	Registry = reg
}

// connectFromURL creates a Redis client from a connection URL string.
// Returns nil if the URL is empty or connection fails.
func connectFromURL(name, url string) *redis.Client {
	if url == "" {
		log.Printf("Redis [%s]: no URL configured, skipping", name)
		return nil
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("Redis [%s]: failed to parse URL: %v", name, err)
		return nil
	}

	client := redis.NewClient(opt)

	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis [%s]: connection failed: %v", name, err)
		return nil
	}

	log.Printf("Redis [%s]: connected successfully", name)
	return client
}
