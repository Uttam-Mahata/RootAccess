package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func ConnectRedis(addr, password string, db int) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Caching will be disabled.", err)
		RDB = nil
		return
	}

	log.Println("Connected to Redis successfully")
}
