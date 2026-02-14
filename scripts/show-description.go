package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("root_access")
	challengesCollection := db.Collection("challenges")

	// Find the challenge
	var challenge bson.M
	err = challengesCollection.FindOne(ctx, bson.M{}).Decode(&challenge)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== FULL DESCRIPTION ===")
	fmt.Println(challenge["description"])
	fmt.Println("\n=== FORMAT ===")
	fmt.Println(challenge["description_format"])
}
