package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var scoreboardCmd = &cobra.Command{
	Use:   "scoreboard",
	Short: "Show the current platform scoreboard",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		client := api.NewClient(cfg)

		// 1. Get active contests to find a contest_id
		var contestsResp struct {
			Contests []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"contests"`
		}
		if err := client.Get("/contests/active", &contestsResp); err != nil {
			fmt.Printf("Error fetching contests: %v\n", err)
			return
		}

		if len(contestsResp.Contests) == 0 {
			fmt.Println("No active contests found.")
			return
		}

		// Use the first active contest
		contestID := contestsResp.Contests[0].ID
		contestName := contestsResp.Contests[0].Name

		fmt.Printf("=== SCOREBOARD: %s ===\n\n", contestName)

		// 2. Get individual scoreboard for that contest
		var scores []struct {
			Username string `json:"username"`
			Score    int    `json:"score"`
			TeamName string `json:"team_name"`
		}

		if err := client.Get("/scoreboard?contest_id="+contestID, &scores); err != nil {
			fmt.Printf("Error fetching scoreboard: %v\n", err)
			return
		}

		if len(scores) == 0 {
			fmt.Println("No scores recorded yet.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "RANK\tUSERNAME\tTEAM\tPOINTS")
		for i, s := range scores {
			team := s.TeamName
			if team == "" {
				team = "-"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%d\n", i+1, s.Username, team, s.Score)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(scoreboardCmd)
}
