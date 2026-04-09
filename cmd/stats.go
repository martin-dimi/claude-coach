package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fridge/coach/internal/db"
	"github.com/spf13/cobra"
)

type StatsResult struct {
	Today    []db.ActivityStat `json:"today"`
	Lifetime []db.ActivityStat `json:"lifetime"`
	Streak   int               `json:"streak"`
	Best     int               `json:"best"`
}

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show your stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := getStats()
			if err != nil {
				return err
			}
			if jsonFlag {
				return json.NewEncoder(os.Stdout).Encode(result)
			}
			renderStatsHuman(result)
			return nil
		},
	}
}

func getStats() (StatsResult, error) {
	now := time.Now().UTC().Add(1 * time.Second)

	today, err := db.Stats(startOfDayUTC(), now)
	if err != nil {
		return StatsResult{}, err
	}

	lifetime, err := db.Stats(time.Time{}, now)
	if err != nil {
		return StatsResult{}, err
	}

	streak, best, err := db.CurrentStreak()
	if err != nil {
		return StatsResult{}, err
	}

	return StatsResult{Today: today, Lifetime: lifetime, Streak: streak, Best: best}, nil
}

func renderStatsHuman(r StatsResult) {
	fmt.Println()
	if len(r.Today) == 0 {
		fmt.Println("  Nothing yet. Get going!")
	} else {
		fmt.Println("  Today")
		for _, s := range r.Today {
			if s.TotalReps > 0 {
				fmt.Printf("    %d %s\n", s.TotalReps, s.Activity)
			} else if s.DoneCount > 0 {
				fmt.Printf("    %dx %s\n", s.DoneCount, s.Activity)
			}
		}
	}
	fmt.Println()
	if len(r.Lifetime) > 0 {
		fmt.Println("  All Time")
		for _, s := range r.Lifetime {
			if s.TotalReps > 0 {
				fmt.Printf("    %d %s\n", s.TotalReps, s.Activity)
			} else if s.DoneCount > 0 {
				fmt.Printf("    %dx %s\n", s.DoneCount, s.Activity)
			}
		}
	}
	fmt.Println()
	fmt.Printf("  %d day streak · best: %d days\n", r.Streak, r.Best)
	fmt.Println()
}
