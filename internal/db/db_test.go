package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	conn = nil
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() { os.Unsetenv("XDG_CONFIG_HOME"); conn = nil })
}

func TestLogAndLastTime(t *testing.T) {
	t.Run("no records returns zero time", func(t *testing.T) {
		setupTestDB(t)
		last, err := LastTime("pushups", "done")
		require.NoError(t, err)
		require.True(t, last.IsZero())
	})

	t.Run("returns time after logging", func(t *testing.T) {
		setupTestDB(t)
		require.NoError(t, LogActivity("pushups", 20, "", "done"))
		last, err := LastTime("pushups", "done")
		require.NoError(t, err)
		require.False(t, last.IsZero())
		require.WithinDuration(t, time.Now(), last, 5*time.Second)
	})

	t.Run("different activity is still zero", func(t *testing.T) {
		setupTestDB(t)
		LogActivity("pushups", 20, "", "done")
		last, err := LastTime("water", "done")
		require.NoError(t, err)
		require.True(t, last.IsZero())
	})
}

func TestLastTimeSeparatesActions(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 0, "", "skip")

	lastDone, _ := LastTime("pushups", "done")
	lastSkip, _ := LastTime("pushups", "skip")

	require.False(t, lastDone.IsZero())
	require.False(t, lastSkip.IsZero())
}

func TestStats(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 0, "", "skip")
	LogActivity("water", 0, "", "done")

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stats, err := Stats(startOfDay, now.Add(1*time.Second))
	require.NoError(t, err)
	require.Len(t, stats, 2)

	pushups := stats[0]
	require.Equal(t, "pushups", pushups.Activity)
	require.Equal(t, 40, pushups.TotalReps)
	require.Equal(t, 2, pushups.DoneCount)
	require.Equal(t, 1, pushups.SkipCount)

	water := stats[1]
	require.Equal(t, "water", water.Activity)
	require.Equal(t, 1, water.DoneCount)
}

func TestStatsTimeRange(t *testing.T) {
	setupTestDB(t)
	LogActivity("pushups", 20, "", "done")
	now := time.Now()

	t.Run("future range returns empty", func(t *testing.T) {
		stats, err := Stats(now.Add(1*time.Hour), now.Add(2*time.Hour))
		require.NoError(t, err)
		require.Empty(t, stats)
	})

	t.Run("lifetime includes all", func(t *testing.T) {
		stats, err := Stats(time.Time{}, now.Add(1*time.Second))
		require.NoError(t, err)
		require.Len(t, stats, 1)
	})
}

func TestStatsByDay(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 20, "", "done")
	LogActivity("water", 0, "", "done")

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stats, err := StatsByDay(startOfDay, now.Add(1*time.Second))
	require.NoError(t, err)
	require.Len(t, stats, 2)

	today := now.Format("2006-01-02")
	for _, s := range stats {
		require.Equal(t, today, s.Date)
	}
}

func TestCurrentStreak(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		setupTestDB(t)
		current, best, err := CurrentStreak()
		require.NoError(t, err)
		require.Equal(t, 0, current)
		require.Equal(t, 0, best)
	})

	t.Run("today only", func(t *testing.T) {
		setupTestDB(t)
		LogActivity("pushups", 20, "", "done")
		current, _, err := CurrentStreak()
		require.NoError(t, err)
		require.Equal(t, 1, current)
	})

	t.Run("consecutive days", func(t *testing.T) {
		setupTestDB(t)
		db, _ := Open()
		today := time.Now().Format("2006-01-02 15:04:05")
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
		db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", today)
		db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", yesterday)

		current, best, err := CurrentStreak()
		require.NoError(t, err)
		require.Equal(t, 2, current)
		require.GreaterOrEqual(t, best, 2)
	})

	t.Run("gap breaks streak", func(t *testing.T) {
		setupTestDB(t)
		db, _ := Open()
		today := time.Now().Format("2006-01-02 15:04:05")
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
		threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02 15:04:05")
		db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", today)
		db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", yesterday)
		db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", threeDaysAgo)

		current, best, _ := CurrentStreak()
		require.Equal(t, 2, current)
		require.Equal(t, 2, best)
	})
}
