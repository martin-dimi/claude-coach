package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fridge/coach/internal/db"
	"github.com/fridge/coach/internal/ui"
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
	now := time.Now().UTC()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	dayStats, _ := db.StatsByDay(thirtyDaysAgo, now.Add(1*time.Second))

	// Group by activity: activity -> date -> DayStat
	byActivity := make(map[string]map[string]db.DayStat)
	var activityOrder []string
	seen := make(map[string]bool)
	for _, s := range dayStats {
		if !seen[s.Activity] {
			seen[s.Activity] = true
			activityOrder = append(activityOrder, s.Activity)
		}
		if byActivity[s.Activity] == nil {
			byActivity[s.Activity] = make(map[string]db.DayStat)
		}
		byActivity[s.Activity][s.Date] = s
	}

	// Also include activities from lifetime that might not have 30-day data
	for _, s := range r.Lifetime {
		if !seen[s.Activity] {
			seen[s.Activity] = true
			activityOrder = append(activityOrder, s.Activity)
			byActivity[s.Activity] = make(map[string]db.DayStat)
		}
	}

	fmt.Println()

	if len(activityOrder) == 0 {
		fmt.Println("  Nothing yet. Get going!")
		fmt.Println()
		return
	}

	// Build today and lifetime lookup
	todayMap := make(map[string]db.ActivityStat)
	for _, s := range r.Today {
		todayMap[s.Activity] = s
	}
	lifetimeMap := make(map[string]db.ActivityStat)
	for _, s := range r.Lifetime {
		lifetimeMap[s.Activity] = s
	}

	for i, activity := range activityOrder {
		today := todayMap[activity]
		lifetime := lifetimeMap[activity]

		grid := ui.RenderGrid(activity, i, byActivity[activity], today.TotalReps, today.DoneCount, lifetime.TotalReps, lifetime.DoneCount)
		fmt.Println(grid)
		fmt.Println()
	}

	fmt.Println(ui.RenderFooter(r.Streak, r.Best))
	fmt.Println()
}
