package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func convertIndentedToFenced(markdown string) string {
	// Pattern to match a line with just a language name followed by indented code
	// Example:
	// C
	//     #include<stdio.h>
	//     int main() { }
	
	lines := strings.Split(markdown, "\n")
	result := []string{}
	i := 0
	
	for i < len(lines) {
		line := lines[i]
		
		// Check if this line looks like a language identifier
		trimmed := strings.TrimSpace(line)
		isLanguage := false
		
		// Common programming language names (short, no spaces)
		commonLangs := []string{"c", "cpp", "python", "java", "javascript", "go", "rust", "ruby", "php", "bash", "shell", "sql", "html", "css", "json", "yaml", "xml"}
		for _, lang := range commonLangs {
			if strings.ToLower(trimmed) == lang {
				isLanguage = true
				break
			}
		}
		
		if isLanguage {
			// Check if next line is indented (code block) or empty then indented
			nextIdx := i + 1
			if nextIdx < len(lines) {
				// Skip one empty line if present
				if strings.TrimSpace(lines[nextIdx]) == "" {
					nextIdx++
				}
				
				if nextIdx < len(lines) && strings.HasPrefix(lines[nextIdx], "    ") {
					// This is a language + indented code block
					language := strings.ToLower(trimmed)
					
					// Skip to first code line
					i = nextIdx
					
					// Add opening fence with language
					result = append(result, "```"+language)
					
					// Collect all indented lines
					for i < len(lines) && (strings.HasPrefix(lines[i], "    ") || strings.TrimSpace(lines[i]) == "") {
						// Remove 4-space indentation
						codeLine := lines[i]
						if strings.HasPrefix(codeLine, "    ") {
							codeLine = codeLine[4:]
						}
						result = append(result, codeLine)
						i++
					}
					
					// Add closing fence
					result = append(result, "```")
					result = append(result, "") // Add blank line after code block
					continue
				}
			}
		}
		
		// Regular line
		result = append(result, line)
		i++
	}
	
	return strings.Join(result, "\n")
}

func convertGenericIndentedBlocks(markdown string) string {
	// Convert remaining indented code blocks (4 spaces) to fenced blocks
	re := regexp.MustCompile(`(?m)^((?:    .+\n?)+)`)
	
	markdown = re.ReplaceAllStringFunc(markdown, func(match string) string {
		lines := strings.Split(strings.TrimSuffix(match, "\n"), "\n")
		var codeLines []string
		for _, line := range lines {
			if strings.HasPrefix(line, "    ") {
				codeLines = append(codeLines, line[4:])
			}
		}
		return "```\n" + strings.Join(codeLines, "\n") + "\n```\n"
	})
	
	return markdown
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
		descriptionFormat := challenge["description_format"]

		if description == nil || descriptionFormat != "markdown" {
			continue
		}

		descStr := fmt.Sprintf("%v", description)
		
		fmt.Printf("Processing: %s\n", title)
		fmt.Printf("Original length: %d chars\n", len(descStr))
		
		// Convert indented code blocks to fenced code blocks
		newDesc := convertIndentedToFenced(descStr)
		newDesc = convertGenericIndentedBlocks(newDesc)
		
		fmt.Printf("New length: %d chars\n", len(newDesc))
		
		if newDesc != descStr {
			fmt.Println("✓ Changes detected, updating...")
			
			// Show a sample of changes
			if len(newDesc) > 500 {
				fmt.Printf("Sample: %s...\n", newDesc[:500])
			}
			
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
