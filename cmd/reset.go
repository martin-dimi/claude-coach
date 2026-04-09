package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/martin-dimi/claude-coach/internal/config"
	"github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Clear all activity data",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := filepath.Join(config.Dir(), "coach.db")
			if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Println("  All data cleared.")
			return nil
		},
	}
}
