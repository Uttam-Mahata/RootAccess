package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rootaccess",
	Short: "RootAccess CTF Platform CLI",
	Long: `RootAccess is a Capture The Flag (CTF) platform CLI tool for managing challenges,
submissions, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root flags if any
}
