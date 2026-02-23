package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var openCmd = &cobra.Command{
	Use:   "open <ID>",
	Short: "View challenge details and description",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.Token == "" {
			fmt.Println("Error: You must be logged in to view challenge details. Run 'rootaccess login' first.")
			return
		}

		client := api.NewClient(cfg)
		var ch struct {
			ID                string   `json:"id"`
			Title             string   `json:"title"`
			Description       string   `json:"description"`
			Category          string   `json:"category"`
			Difficulty        string   `json:"difficulty"`
			CurrentPoints     int      `json:"current_points"`
			Files             []string `json:"files"`
			Tags              []string `json:"tags"`
			IsSolved          bool     `json:"is_solved"`
		}

		if err := client.Get("/challenges/"+id, &ch); err != nil {
			fmt.Printf("Error fetching challenge details: %v\n", err)
			return
		}

		fmt.Printf("\n=== %s ===\n", strings.ToUpper(ch.Title))
		fmt.Printf("ID:         %s\n", ch.ID)
		fmt.Printf("Category:   %s\n", ch.Category)
		fmt.Printf("Difficulty: %s\n", ch.Difficulty)
		fmt.Printf("Points:     %d\n", ch.CurrentPoints)
		if ch.IsSolved {
			fmt.Printf("Status:     SOLVED ●\n")
		} else {
			fmt.Printf("Status:     UNSOLVED ○\n")
		}
		if len(ch.Tags) > 0 {
			fmt.Printf("Tags:       %s\n", strings.Join(ch.Tags, ", "))
		}
		fmt.Printf("\nDESCRIPTION:\n%s\n", ch.Description)

		if len(ch.Files) > 0 {
			fmt.Printf("\nFILES:\n")
			for _, file := range ch.Files {
				fmt.Printf("- %s\n", file)
			}
		}

		fmt.Printf("\nSubmit flag:\n  rootaccess submit %s <FLAG>\n", ch.ID)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
