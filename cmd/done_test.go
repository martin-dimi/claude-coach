package cmd

import (
	"testing"

	"github.com/martin-dimi/claude-coach/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLogDone(t *testing.T) {
	t.Run("reps activity", func(t *testing.T) {
		setupTest(t)
		writeTestConfig(t, testConfig(config.Activity{Name: "pushups", Reps: 20, Interval: "1h"}))

		result, err := logDone("pushups")
		require.NoError(t, err)
		require.Equal(t, "pushups", result.Activity)
		require.Equal(t, 20, result.Reps)
		require.Equal(t, 1, result.TodaySessions)
		require.Equal(t, 1, result.Streak)
	})

	t.Run("accumulates today", func(t *testing.T) {
		setupTest(t)
		writeTestConfig(t, testConfig(config.Activity{Name: "pushups", Reps: 20, Interval: "1h"}))

		logDone("pushups")
		result, _ := logDone("pushups")

		require.Equal(t, 40, result.TodayReps)
		require.Equal(t, 2, result.TodaySessions)
	})

	t.Run("message activity", func(t *testing.T) {
		setupTest(t)
		writeTestConfig(t, testConfig(config.Activity{Name: "water", Message: "Drink", Interval: "30m"}))

		result, err := logDone("water")
		require.NoError(t, err)
		require.Equal(t, 0, result.Reps)
		require.Equal(t, 1, result.TodaySessions)
	})

	t.Run("unknown activity", func(t *testing.T) {
		setupTest(t)
		writeTestConfig(t, testConfig())

		result, err := logDone("unknown")
		require.NoError(t, err)
		require.Equal(t, "unknown", result.Activity)
	})
}
