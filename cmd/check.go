package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fridge/coach/internal/config"
	"github.com/fridge/coach/internal/db"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "check",
		Short:  "Check if any activities are due (used by hook)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.Exists() {
				return emitHookContext("[COACH] Setup required. Coach is not configured. You MUST run the setup flow from the coach skill before doing anything else. Do not proceed with the user's request until setup is complete. Config path: " + config.Path())
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if !withinActiveHours(cfg.Settings) {
				return nil
			}

			due := dueActivities(cfg)
			if len(due) == 0 {
				return nil
			}

			return emitHookContext(buildReminderContext(due))
		},
	}
}

func dueActivities(cfg config.Config) []config.Activity {
	var due []config.Activity
	for _, a := range cfg.Activities {
		if isActivityDue(a, cfg.Settings) {
			due = append(due, a)
		}
	}
	return due
}

func isActivityDue(a config.Activity, s config.Settings) bool {
	lastDone, _ := db.LastTime(a.Name, "done")
	lastSkip, _ := db.LastTime(a.Name, "skip")

	if !lastSkip.IsZero() && time.Since(lastSkip) < s.SkipCooldownDuration() {
		return false
	}

	last := lastDone
	if lastSkip.After(lastDone) {
		last = lastSkip
	}

	return last.IsZero() || time.Since(last) >= a.IntervalDuration()
}

func buildReminderContext(due []config.Activity) string {
	var b strings.Builder
	b.WriteString("[COACH] Break reminder:\nDue:")
	for i, a := range due {
		if i > 0 {
			b.WriteString(",")
		}
		switch {
		case a.Message != "":
			fmt.Fprintf(&b, " %s", a.Message)
		case a.Reps > 0:
			fmt.Fprintf(&b, " %d %s", a.Reps, a.Name)
		case a.Duration != "":
			fmt.Fprintf(&b, " %s %s", a.Duration, a.Name)
		default:
			fmt.Fprintf(&b, " %s", a.Name)
		}
	}

	if stats := todayStats(); len(stats) > 0 {
		var parts []string
		for _, s := range stats {
			if s.TotalReps > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", s.TotalReps, s.Activity))
			} else if s.DoneCount > 0 {
				parts = append(parts, fmt.Sprintf("%dx %s", s.DoneCount, s.Activity))
			}
		}
		if len(parts) > 0 {
			fmt.Fprintf(&b, "\nToday: %s", strings.Join(parts, ", "))
		}
	}

	if streak, _, _ := db.CurrentStreak(); streak > 0 {
		if streak == 1 {
			b.WriteString("\nStreak: 1 day")
		} else {
			fmt.Fprintf(&b, "\nStreak: %d days", streak)
		}
	}

	b.WriteString("\nLog with: coach done <activity> --json | Skip with: coach skip <activity> --json")
	return b.String()
}

func emitHookContext(context string) error {
	// Output as both plain text (visible in transcript) and
	// the context itself is directive enough for Claude to act on.
	_, err := fmt.Fprintln(os.Stdout, context)
	return err
}
