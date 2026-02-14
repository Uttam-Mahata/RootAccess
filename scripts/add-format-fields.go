package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Get MongoDB connection string from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	
	dbName := os.Getenv("MONGO_DATABASE")
	if dbName == "" {
		dbName = "ctf_platform"
	}

	fmt.Printf("Connecting to MongoDB: %s\n", mongoURI)
	fmt.Printf("Database: %s\n", dbName)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	db := client.Database(dbName)

	// Migrate challenges collection
	fmt.Println("\n=== Migrating Challenges Collection ===")
	challengesCollection := db.Collection("challenges")
	
	// Add description_format field to all challenges without it
	challengeFilter := bson.M{
		"$or": []bson.M{
			{"description_format": bson.M{"$exists": false}},
			{"description_format": ""},
		},
	}
	challengeUpdate := bson.M{
		"$set": bson.M{
			"description_format": "markdown",
		},
	}
	
	challengeResult, err := challengesCollection.UpdateMany(ctx, challengeFilter, challengeUpdate)
	if err != nil {
		log.Fatal("Failed to update challenges:", err)
	}
	fmt.Printf("✓ Updated %d challenges with description_format='markdown'\n", challengeResult.ModifiedCount)

	// Migrate writeups collection
	fmt.Println("\n=== Migrating Writeups Collection ===")
	writeupsCollection := db.Collection("writeups")
	
	// Add content_format field to all writeups without it
	writeupFilter := bson.M{
		"$or": []bson.M{
			{"content_format": bson.M{"$exists": false}},
			{"content_format": ""},
		},
	}
	writeupUpdate := bson.M{
		"$set": bson.M{
			"content_format": "markdown",
		},
	}
	
	writeupResult, err := writeupsCollection.UpdateMany(ctx, writeupFilter, writeupUpdate)
	if err != nil {
		log.Fatal("Failed to update writeups:", err)
	}
	fmt.Printf("✓ Updated %d writeups with content_format='markdown'\n", writeupResult.ModifiedCount)

	// Summary
	fmt.Println("\n=== Migration Complete ===")
	fmt.Printf("Total challenges migrated: %d\n", challengeResult.ModifiedCount)
	fmt.Printf("Total writeups migrated: %d\n", writeupResult.ModifiedCount)
	fmt.Println("\nAll existing records now have format fields set to 'markdown' by default.")
	fmt.Println("New records will require explicit format specification from the frontend.")
}
