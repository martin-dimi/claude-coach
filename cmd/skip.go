package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fridge/coach/internal/config"
	"github.com/fridge/coach/internal/db"
	"github.com/spf13/cobra"
)

type skipOutput struct {
	Activity string `json:"activity"`
	Skipped  bool   `json:"skipped"`
	Cooldown string `json:"cooldown"`
}

var skipCmd = &cobra.Command{
	Use:   "skip <activity>",
	Short: "Skip an activity (will remind again after cooldown)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		activityName := args[0]

		if err := db.LogActivity(activityName, 0, "", "skip"); err != nil {
			return err
		}

		cfg, _ := config.Load()
		cooldown := cfg.Settings.SkipCooldown
		if cooldown == "" {
			cooldown = "10m"
		}

		out := skipOutput{
			Activity: activityName,
			Skipped:  true,
			Cooldown: cooldown,
		}

		if jsonFlag {
			return json.NewEncoder(os.Stdout).Encode(out)
		}

		fmt.Printf("  Skipped %s. Reminder in %s.\n", activityName, cooldown)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skipCmd)
}
