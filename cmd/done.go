package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fridge/coach/internal/config"
	"github.com/fridge/coach/internal/db"
	"github.com/spf13/cobra"
)

type doneOutput struct {
	Activity string `json:"activity"`
	Reps     int    `json:"reps,omitempty"`
	Duration string `json:"duration,omitempty"`
	Today    int    `json:"today"`
	TodayX   int    `json:"today_sessions"`
	Streak   int    `json:"streak"`
	AllTime  int    `json:"all_time"`
}

var doneCmd = &cobra.Command{
	Use:   "done <activity>",
	Short: "Log that you completed an activity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		activityName := args[0]

		// Find activity in config for reps/duration
		var reps int
		var duration string
		cfg, err := config.Load()
		if err == nil {
			for _, a := range cfg.Activities {
				if a.Name == activityName {
					reps = a.Reps
					duration = a.Duration
					break
				}
			}
		}

		if err := db.LogActivity(activityName, reps, duration, "done"); err != nil {
			return err
		}

		// Get today stats for this activity
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		todayStats, _ := db.Stats(startOfDay, now)

		var todayReps, todaySessions int
		for _, s := range todayStats {
			if s.Activity == activityName {
				todayReps = s.TotalReps
				todaySessions = s.DoneCount
				break
			}
		}

		streak, _, _ := db.CurrentStreak()

		// All time for this activity
		allStats, _ := db.Stats(time.Time{}, now)
		var allTime int
		for _, s := range allStats {
			if s.Activity == activityName {
				if s.TotalReps > 0 {
					allTime = s.TotalReps
				} else {
					allTime = s.DoneCount
				}
				break
			}
		}

		out := doneOutput{
			Activity: activityName,
			Reps:     reps,
			Duration: duration,
			Today:    todayReps,
			TodayX:   todaySessions,
			Streak:   streak,
			AllTime:  allTime,
		}

		if jsonFlag {
			return json.NewEncoder(os.Stdout).Encode(out)
		}

		// Human mode
		if reps > 0 {
			fmt.Printf("  ✓ %d %s · %d today · %d day streak\n", reps, activityName, todayReps, streak)
		} else {
			fmt.Printf("  ✓ %s · %dx today · %d day streak\n", activityName, todaySessions, streak)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
