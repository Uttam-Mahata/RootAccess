package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB(uri, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	// Configure connection pooling
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMaxPoolSize(100)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	DB = client.Database(dbName)
	log.Println("Connected to MongoDB successfully")
}

func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}
