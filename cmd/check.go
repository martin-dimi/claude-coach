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

type checkOutput struct {
	HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
}

type hookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

type checkDueActivity struct {
	Name     string `json:"name"`
	Reps     int    `json:"reps,omitempty"`
	Duration string `json:"duration,omitempty"`
	Message  string `json:"message,omitempty"`
}

var checkCmd = &cobra.Command{
	Use:    "check",
	Short:  "Check if any activities are due (used by hook)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// No config = guide setup
		if !config.Exists() {
			return outputHook("[COACH] Not configured. Guide the user through setup. Write config to " + config.Path() + ". See skill for config schema and presets.")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Check active hours
		if !withinActiveHours(cfg.Settings) {
			return nil
		}

		// Check each activity
		var due []checkDueActivity
		for _, a := range cfg.Activities {
			lastDone, _ := db.LastTime(a.Name, "done")
			lastSkip, _ := db.LastTime(a.Name, "skip")

			// Check skip cooldown
			if !lastSkip.IsZero() {
				cooldown := cfg.Settings.SkipCooldownDuration()
				if time.Since(lastSkip) < cooldown {
					continue
				}
			}

			// Use the most recent done or skip as the timer reference
			lastActivity := lastDone
			if lastSkip.After(lastDone) {
				lastActivity = lastSkip
			}

			// If no log at all, activity is due (first reminder after setup)
			if lastActivity.IsZero() || time.Since(lastActivity) >= a.IntervalDuration() {
				da := checkDueActivity{
					Name:     a.Name,
					Reps:     a.Reps,
					Duration: a.Duration,
					Message:  a.Message,
				}
				due = append(due, da)
			}
		}

		if len(due) == 0 {
			return nil // nothing due, no output
		}

		// Build context string
		ctx := "[COACH] Break reminder:\nDue:"
		for _, d := range due {
			if d.Message != "" {
				ctx += fmt.Sprintf(" %s,", d.Message)
			} else if d.Reps > 0 {
				ctx += fmt.Sprintf(" %d %s,", d.Reps, d.Name)
			} else if d.Duration != "" {
				ctx += fmt.Sprintf(" %s %s,", d.Duration, d.Name)
			} else {
				ctx += fmt.Sprintf(" %s,", d.Name)
			}
		}
		ctx = ctx[:len(ctx)-1] // trim trailing comma

		// Add today stats
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		todayStats, _ := db.Stats(startOfDay, now)
		if len(todayStats) > 0 {
			ctx += "\nToday:"
			for _, s := range todayStats {
				if s.TotalReps > 0 {
					ctx += fmt.Sprintf(" %d %s,", s.TotalReps, s.Activity)
				} else if s.DoneCount > 0 {
					ctx += fmt.Sprintf(" %dx %s,", s.DoneCount, s.Activity)
				}
			}
			ctx = ctx[:len(ctx)-1]
		}

		// Add streak
		streak, _, _ := db.CurrentStreak()
		if streak > 0 {
			ctx += fmt.Sprintf("\nStreak: %d days", streak)
		}

		ctx += "\nLog with: coach done <activity> --json | Skip with: coach skip <activity> --json"

		return outputHook(ctx)
	},
}

func outputHook(context string) error {
	out := checkOutput{
		HookSpecificOutput: hookOutput{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: context,
		},
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func withinActiveHours(s config.Settings) bool {
	now := time.Now()
	start, err1 := time.Parse("15:04", s.ActiveHours[0])
	end, err2 := time.Parse("15:04", s.ActiveHours[1])
	if err1 != nil || err2 != nil {
		return true
	}
	nowMin := now.Hour()*60 + now.Minute()
	startMin := start.Hour()*60 + start.Minute()
	endMin := end.Hour()*60 + end.Minute()
	return nowMin >= startMin && nowMin <= endMin
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
