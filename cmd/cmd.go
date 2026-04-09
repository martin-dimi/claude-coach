package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/martin-dimi/claude-coach/internal/config"
	"github.com/martin-dimi/claude-coach/internal/db"
	"github.com/spf13/cobra"
)

var jsonFlag bool

var rootCmd = &cobra.Command{
	Use:   "coach",
	Short: "Your personal wellness coach inside Claude Code",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON (for Claude)")

	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newDoneCmd())
	rootCmd.AddCommand(newSkipCmd())
	rootCmd.AddCommand(newStatsCmd())
	rootCmd.AddCommand(newResetCmd())
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Shared helpers

func startOfDayUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func todayStats() []db.ActivityStat {
	stats, _ := db.Stats(startOfDayUTC(), time.Now().UTC().Add(1*time.Second))
	return stats
}

func findActivity(cfg config.Config, name string) (config.Activity, bool) {
	for _, a := range cfg.Activities {
		if a.Name == name {
			return a, true
		}
	}
	return config.Activity{}, false
}

func withinActiveHours(s config.Settings) bool {
	now := time.Now()
	start, err1 := time.Parse("15:04", s.ActiveHours[0])
	end, err2 := time.Parse("15:04", s.ActiveHours[1])
	if err1 != nil || err2 != nil {
		return true
	}
	nowMin := now.Hour()*60 + now.Minute()
	return nowMin >= start.Hour()*60+start.Minute() && nowMin <= end.Hour()*60+end.Minute()
}
