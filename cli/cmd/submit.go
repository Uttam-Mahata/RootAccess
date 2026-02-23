package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var submitCmd = &cobra.Command{
	Use:   "submit <ID> <FLAG>",
	Short: "Submit a flag for a challenge",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		flag := args[1]

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.Token == "" {
			fmt.Println("Error: You must be logged in to submit flags. Run 'rootaccess login' first.")
			return
		}

		client := api.NewClient(cfg)
		body := map[string]string{"flag": flag}
		var resp struct {
			Correct       bool   `json:"correct"`
			AlreadySolved bool   `json:"already_solved"`
			Message       string `json:"message"`
			Points        int    `json:"points"`
		}

		if err := client.Post("/challenges/"+id+"/submit", body, &resp); err != nil {
			fmt.Printf("Error submitting flag: %v\n", err)
			return
		}

		if resp.Correct {
			if resp.AlreadySolved {
				fmt.Printf("● Correct! But you already solved this challenge.\n")
			} else {
				fmt.Printf("● CORRECT! Flag is valid. You earned %d points!\n", resp.Points)
				if resp.Message != "" {
					fmt.Printf("%s\n", resp.Message)
				}
			}
		} else {
			fmt.Printf("○ INCORRECT. Try again!\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
