package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently logged in user",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.Token == "" {
			fmt.Println("You are not logged in. Run 'rootaccess login' first.")
			return
		}

		client := api.NewClient(cfg)
		var userResp struct {
			Authenticated bool `json:"authenticated"`
			User          struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				Role     string `json:"role"`
			} `json:"user"`
		}

		if err := client.Get("/auth/me", &userResp); err != nil {
			fmt.Printf("Error checking authentication: %v\n", err)
			return
		}

		if !userResp.Authenticated {
			fmt.Println("Error: Session is invalid or expired. Please login again.")
			return
		}

		fmt.Printf("Logged in as:\n")
		fmt.Printf("Username: %s\n", userResp.User.Username)
		fmt.Printf("Email:    %s\n", userResp.User.Email)
		fmt.Printf("Role:     %s\n", userResp.User.Role)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
