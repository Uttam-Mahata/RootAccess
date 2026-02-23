package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to RootAccess using Google OAuth",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		loginURL := fmt.Sprintf("%s/auth/google?cli=true", cfg.BaseURL)
		fmt.Printf("Opening browser for Google Login...\n")
		fmt.Printf("If the browser doesn't open automatically, visit: %s\n", loginURL)

		if err := open.Run(loginURL); err != nil {
			fmt.Printf("Could not open browser: %v\n", err)
		}

		fmt.Printf("\nAfter logging in, please paste your access token here:\n")
		fmt.Printf("Token: ")

		reader := bufio.NewReader(os.Stdin)
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		if token == "" {
			fmt.Println("Error: Token cannot be empty.")
			return
		}

		cfg.Token = token

		// Verify token by calling /auth/me
		client := api.NewClient(cfg)
		var userResp struct {
			Authenticated bool `json:"authenticated"`
			User          struct {
				Username string `json:"username"`
			} `json:"user"`
		}

		if err := client.Get("/auth/me", &userResp); err != nil {
			fmt.Printf("Error: Authentication failed: %v\n", err)
			return
		}

		if !userResp.Authenticated {
			fmt.Println("Error: Token is invalid or expired.")
			return
		}

		cfg.Username = userResp.User.Username
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Successfully logged in as %s!\n", cfg.Username)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
