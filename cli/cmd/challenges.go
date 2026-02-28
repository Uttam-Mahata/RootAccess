package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var challengesCmd = &cobra.Command{
	Use:   "challenges",
	Short: "List all available challenges",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.Token == "" {
			fmt.Println("Error: You must be logged in to view challenges. Run 'rootaccess login' first.")
			return
		}

		client := api.NewClient(cfg)
		var challenges []struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			Category      string `json:"category"`
			Difficulty    string `json:"difficulty"`
			CurrentPoints int    `json:"current_points"`
			SolveCount    int    `json:"solve_count"`
			IsSolved      bool   `json:"is_solved"`
		}

		if err := client.Get("/challenges", &challenges); err != nil {
			fmt.Printf("Error fetching challenges: %v\n", err)
			return
		}

		if len(challenges) == 0 {
			fmt.Println("No challenges available at the moment.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tPOINTS\tSOLVES\tSTATUS")
		for _, ch := range challenges {
			status := "○"
			if ch.IsSolved {
				status = "● SOLVED"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n", ch.ID, ch.Title, ch.Category, ch.CurrentPoints, ch.SolveCount, status)
		}
		w.Flush()
		
		fmt.Printf("\nUse 'rootaccess open <ID>' to view challenge details.\n")
	},
}

func init() {
	rootCmd.AddCommand(challengesCmd)
}
