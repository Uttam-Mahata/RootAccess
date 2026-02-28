package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.Token == "" {
			fmt.Println("You are not logged in.")
			return
		}

		if err := cfg.Clear(); err != nil {
			fmt.Printf("Error logging out: %v\n", err)
			return
		}

		fmt.Println("Logged out successfully.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
