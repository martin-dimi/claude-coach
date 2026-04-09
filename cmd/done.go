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

type DoneResult struct {
	Activity      string `json:"activity"`
	Reps          int    `json:"reps,omitempty"`
	Duration      string `json:"duration,omitempty"`
	TodayReps     int    `json:"today_reps"`
	TodaySessions int    `json:"today_sessions"`
	Streak        int    `json:"streak"`
	AllTime       int    `json:"all_time"`
}

func newDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done <activity>",
		Short: "Log that you completed an activity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := logDone(args[0])
			if err != nil {
				return err
			}
			if jsonFlag {
				return json.NewEncoder(os.Stdout).Encode(result)
			}
			if result.Reps > 0 {
				fmt.Printf("  ✓ %d %s · %d today · %d day streak\n", result.Reps, result.Activity, result.TodayReps, result.Streak)
			} else {
				fmt.Printf("  ✓ %s · %dx today · %d day streak\n", result.Activity, result.TodaySessions, result.Streak)
			}
			return nil
		},
	}
}

func logDone(name string) (DoneResult, error) {
	cfg, _ := config.Load()
	a, _ := findActivity(cfg, name)

	if err := db.LogActivity(name, a.Reps, a.Duration, "done"); err != nil {
		return DoneResult{}, err
	}

	var todayReps, todaySessions int
	for _, s := range todayStats() {
		if s.Activity == name {
			todayReps = s.TotalReps
			todaySessions = s.DoneCount
			break
		}
	}

	streak, _, _ := db.CurrentStreak()

	var allTime int
	allStats, _ := db.Stats(time.Time{}, time.Now().UTC().Add(1*time.Second))
	for _, s := range allStats {
		if s.Activity == name {
			allTime = max(s.TotalReps, s.DoneCount)
			break
		}
	}

	return DoneResult{
		Activity:      name,
		Reps:          a.Reps,
		Duration:      a.Duration,
		TodayReps:     todayReps,
		TodaySessions: todaySessions,
		Streak:        streak,
		AllTime:       allTime,
	}, nil
}
