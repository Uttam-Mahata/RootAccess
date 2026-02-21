package database

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func ConnectRedis(addr, password string, db int) {
	options := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	// Enable TLS for AWS ElastiCache or any secured Redis
	if password != "" {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	RDB = redis.NewClient(options)

	ctx := context.Background()
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis at %s: %v. Caching will be disabled.", addr, err)
		RDB = nil
		return
	}

	log.Println("Connected to Redis successfully")
}
