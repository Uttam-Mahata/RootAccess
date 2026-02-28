package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func removeStandaloneLangBeforeFence(markdown string) string {
	lines := strings.Split(markdown, "\n")
	result := []string{}
	
	commonLangs := map[string]bool{
		"c": true, "cpp": true, "c++": true, "python": true, "py": true, 
		"java": true, "javascript": true, "js": true, "go": true, "golang": true,
		"rust": true, "ruby": true, "php": true, "bash": true, "shell": true, 
		"sh": true, "sql": true, "html": true, "css": true, "json": true, 
		"yaml": true, "yml": true, "xml": true, "typescript": true, "ts": true,
	}
	
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lowerTrimmed := strings.ToLower(trimmed)
		
		// Check if this is a standalone language identifier
		if commonLangs[lowerTrimmed] {
			// Check if next line (or after blank line) starts with ```
			nextIdx := i + 1
			if nextIdx < len(lines) {
				nextLine := strings.TrimSpace(lines[nextIdx])
				
				// Skip one blank line if present
				if nextLine == "" && nextIdx+1 < len(lines) {
					nextIdx++
					nextLine = strings.TrimSpace(lines[nextIdx])
				}
				
				// If next line starts with ```, skip the language line
				if strings.HasPrefix(nextLine, "```") {
					fmt.Printf("Removing standalone language identifier: '%s'\n", trimmed)
					i++ // Skip this line
					continue
				}
			}
		}
		
		result = append(result, line)
		i++
	}
	
	return strings.Join(result, "\n")
}

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

		if description == nil {
			continue
		}

		descStr := fmt.Sprintf("%v", description)
		
		fmt.Printf("Processing: %s\n", title)
		
		// Remove standalone language identifiers before fenced code blocks
		newDesc := removeStandaloneLangBeforeFence(descStr)
		
		if newDesc != descStr {
			fmt.Println("✓ Changes detected, updating...")
			
			update := bson.M{
				"$set": bson.M{
					"description": newDesc,
				},
			}
			
			_, err := challengesCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
			if err != nil {
				fmt.Printf("❌ Error updating: %v\n", err)
			} else {
				fmt.Println("✅ Updated successfully")
			}
		} else {
			fmt.Println("- No changes needed")
		}
		
		fmt.Println()
	}

	fmt.Println("Done!")
}
