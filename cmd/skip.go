package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fridge/coach/internal/config"
	"github.com/fridge/coach/internal/db"
	"github.com/spf13/cobra"
)

type SkipResult struct {
	Activity string `json:"activity"`
	Skipped  bool   `json:"skipped"`
	Cooldown string `json:"cooldown"`
}

func newSkipCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "skip <activity>",
		Short: "Skip an activity (will remind again after cooldown)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := logSkip(args[0])
			if err != nil {
				return err
			}
			if jsonFlag {
				return json.NewEncoder(os.Stdout).Encode(result)
			}
			fmt.Printf("  Skipped %s. Reminder in %s.\n", result.Activity, result.Cooldown)
			return nil
		},
	}
}

func logSkip(name string) (SkipResult, error) {
	if err := db.LogActivity(name, 0, "", "skip"); err != nil {
		return SkipResult{}, err
	}

	cfg, _ := config.Load()
	cooldown := cfg.Settings.SkipCooldown
	if cooldown == "" {
		cooldown = "10m"
	}

	return SkipResult{Activity: name, Skipped: true, Cooldown: cooldown}, nil
}
