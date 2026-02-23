package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/api"
	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to RootAccess using your browser",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// 1. Find an available port for the callback server
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("Error starting local server: %v\n", err)
			return
		}
		port := listener.Addr().(*net.TCPAddr).Port

		// 2. Start a local HTTP server to receive the token
		tokenChan := make(chan string)
		server := &http.Server{}

		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token != "" {
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				fmt.Fprint(w, "<html><body style='font-family:sans-serif; text-align:center; padding-top:50px;'><h1>Login Successful!</h1><p>You can close this window and return to your terminal.</p></body></html>")
				tokenChan <- token
			} else {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, "Token missing")
			}
		})

		go func() {
			server.Serve(listener)
		}()

		// 3. Open the browser to the frontend CLI auth page
		// Determine frontend URL from BaseURL (strip /api)
		frontendURL := strings.TrimSuffix(cfg.BaseURL, "/api")
		loginURL := fmt.Sprintf("%s/cli/auth?port=%d", frontendURL, port)

		fmt.Printf("Opening your browser to authenticate...\n")
		fmt.Printf("If the browser doesn't open automatically, visit: %s\n", loginURL)

		if err := open.Run(loginURL); err != nil {
			fmt.Printf("Could not open browser: %v\n", err)
		}

		// 4. Wait for the token with a timeout
		fmt.Println("Waiting for authentication...")
		select {
		case token := <-tokenChan:
			cfg.Token = token
			// Verify token
			client := api.NewClient(cfg)
			var userResp struct {
				Authenticated bool `json:"authenticated"`
				User          struct {
					Username string `json:"username"`
				} `json:"user"`
			}

			if err := client.Get("/auth/me", &userResp); err != nil {
				fmt.Printf("Error: Authentication failed: %v\n", err)
			} else if !userResp.Authenticated {
				fmt.Println("Error: Received token is invalid.")
			} else {
				cfg.Username = userResp.User.Username
				if err := cfg.Save(); err != nil {
					fmt.Printf("Error saving config: %v\n", err)
				} else {
					fmt.Printf("Successfully logged in as %s!\n", cfg.Username)
				}
			}
		case <-time.After(5 * time.Minute):
			fmt.Println("Error: Login timed out after 5 minutes.")
		}

		// Shutdown the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
