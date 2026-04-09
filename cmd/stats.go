package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fridge/coach/internal/db"
	"github.com/spf13/cobra"
)

type statsJSON struct {
	Today    []db.ActivityStat `json:"today"`
	Streak   int               `json:"streak"`
	Best     int               `json:"best"`
	Lifetime []db.ActivityStat `json:"lifetime"`
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show your stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		todayStats, err := db.Stats(startOfDay, now)
		if err != nil {
			return err
		}

		lifetimeStats, err := db.Stats(time.Time{}, now)
		if err != nil {
			return err
		}

		streak, best, err := db.CurrentStreak()
		if err != nil {
			return err
		}

		if jsonFlag {
			return json.NewEncoder(os.Stdout).Encode(statsJSON{
				Today:    todayStats,
				Streak:   streak,
				Best:     best,
				Lifetime: lifetimeStats,
			})
		}

		// Human mode - contribution grids will be added in Phase 5
		fmt.Println()

		if len(todayStats) == 0 {
			fmt.Println("  Nothing yet. Get going!")
		} else {
			fmt.Println("  Today")
			for _, s := range todayStats {
				if s.TotalReps > 0 {
					fmt.Printf("    %d %s\n", s.TotalReps, s.Activity)
				} else if s.DoneCount > 0 {
					fmt.Printf("    %dx %s\n", s.DoneCount, s.Activity)
				}
			}
		}

		fmt.Println()

		if len(lifetimeStats) > 0 {
			fmt.Println("  All Time")
			for _, s := range lifetimeStats {
				if s.TotalReps > 0 {
					fmt.Printf("    %d %s\n", s.TotalReps, s.Activity)
				} else if s.DoneCount > 0 {
					fmt.Printf("    %dx %s\n", s.DoneCount, s.Activity)
				}
			}
		}

		fmt.Println()
		fmt.Printf("  %d day streak · best: %d days\n", streak, best)
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
