package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/fridge/coach/internal/config"
	"github.com/fridge/coach/internal/db"
	"github.com/stretchr/testify/require"
)

func TestIsActivityDue(t *testing.T) {
	s := config.Settings{SkipCooldown: "10m"}
	pushups := config.Activity{Name: "pushups", Reps: 20, Interval: "1h"}

	t.Run("no prior log", func(t *testing.T) {
		setupTest(t)
		require.True(t, isActivityDue(pushups, s))
	})

	t.Run("recently done", func(t *testing.T) {
		setupTest(t)
		db.LogActivity("pushups", 20, "", "done")
		require.False(t, isActivityDue(pushups, s))
	})

	t.Run("interval elapsed", func(t *testing.T) {
		setupTest(t)
		db.LogActivity("pushups", 20, "", "done")
		conn, _ := db.Open()
		conn.Exec("UPDATE activity_log SET created_at = ?",
			time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
		require.True(t, isActivityDue(pushups, s))
	})

	t.Run("skip cooldown active", func(t *testing.T) {
		setupTest(t)
		db.LogActivity("pushups", 0, "", "skip")
		require.False(t, isActivityDue(pushups, s))
	})

	t.Run("skip cooldown expired", func(t *testing.T) {
		setupTest(t)
		db.LogActivity("pushups", 0, "", "skip")
		conn, _ := db.Open()
		conn.Exec("UPDATE activity_log SET created_at = ?",
			time.Now().UTC().Add(-2*time.Hour).Format("2006-01-02 15:04:05"))
		require.True(t, isActivityDue(pushups, s))
	})
}

func TestDueActivities(t *testing.T) {
	t.Run("all due when no log", func(t *testing.T) {
		setupTest(t)
		cfg := testConfig(
			config.Activity{Name: "pushups", Reps: 20, Interval: "1h"},
			config.Activity{Name: "water", Message: "Drink water", Interval: "30m"},
		)
		require.Len(t, dueActivities(cfg), 2)
	})

	t.Run("none due after completion", func(t *testing.T) {
		setupTest(t)
		cfg := testConfig(
			config.Activity{Name: "pushups", Reps: 20, Interval: "1h"},
			config.Activity{Name: "water", Message: "Drink water", Interval: "30m"},
		)
		db.LogActivity("pushups", 20, "", "done")
		db.LogActivity("water", 0, "", "done")
		require.Empty(t, dueActivities(cfg))
	})

	t.Run("partial due", func(t *testing.T) {
		setupTest(t)
		cfg := testConfig(
			config.Activity{Name: "pushups", Reps: 20, Interval: "1h"},
			config.Activity{Name: "water", Message: "Drink water", Interval: "30m"},
		)
		db.LogActivity("pushups", 20, "", "done")

		due := dueActivities(cfg)
		require.Len(t, due, 1)
		require.Equal(t, "water", due[0].Name)
	})
}

func TestBuildReminderContext(t *testing.T) {
	tests := []struct {
		name     string
		due      []config.Activity
		contains string
	}{
		{"reps activity", []config.Activity{{Name: "pushups", Reps: 20}}, "20 pushups"},
		{"message activity", []config.Activity{{Name: "water", Message: "Drink a glass of water"}}, "Drink a glass of water"},
		{"duration activity", []config.Activity{{Name: "stretch", Duration: "2m"}}, "2m stretch"},
		{"has COACH prefix", []config.Activity{{Name: "pushups", Reps: 20}}, "[COACH]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			ctx := buildReminderContext(tt.due)
			require.True(t, strings.Contains(ctx, tt.contains), "expected %q in:\n%s", tt.contains, ctx)
		})
	}
}
