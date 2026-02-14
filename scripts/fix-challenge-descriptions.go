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

	// Find all challenges
	cursor, err := challengesCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var challenges []bson.M
	if err = cursor.All(ctx, &challenges); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d challenges\n\n", len(challenges))

	for _, challenge := range challenges {
		id := challenge["_id"]
		title := challenge["title"]
		description := challenge["description"]
		descriptionFormat := challenge["description_format"]

		fmt.Printf("Challenge: %s (ID: %v)\n", title, id)
		fmt.Printf("  Format: %v\n", descriptionFormat)
		
		if description != nil {
			descStr := fmt.Sprintf("%v", description)
			fmt.Printf("  Description length: %d chars\n", len(descStr))
			
			if len(descStr) > 100 {
				fmt.Printf("  Preview: %s...\n", descStr[:100])
			} else {
				fmt.Printf("  Preview: %s\n", descStr)
			}
		} else {
			fmt.Println("  Description: <nil>")
		}

		// Check if description_format is missing
		if descriptionFormat == nil {
			fmt.Println("  ⚠️  WARNING: description_format is missing! Setting to 'markdown'")
			
			update := bson.M{
				"$set": bson.M{
					"description_format": "markdown",
				},
			}
			
			_, err := challengesCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
			if err != nil {
				fmt.Printf("  ❌ Error updating: %v\n", err)
			} else {
				fmt.Println("  ✅ Updated successfully")
			}
		}

		fmt.Println()
	}

	fmt.Println("Done!")
}
