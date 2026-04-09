package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var jsonFlag bool

var rootCmd = &cobra.Command{
	Use:   "coach",
	Short: "Your personal wellness coach inside Claude Code",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON (for Claude)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
