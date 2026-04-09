package cmd

import (
	"testing"

	"github.com/martin-dimi/claude-coach/internal/db"
	"github.com/stretchr/testify/require"
)

func TestGetStats(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		setupTest(t)
		result, err := getStats()
		require.NoError(t, err)
		require.Empty(t, result.Today)
		require.Equal(t, 0, result.Streak)
	})

	t.Run("with data", func(t *testing.T) {
		setupTest(t)
		db.LogActivity("pushups", 20, "", "done")
		db.LogActivity("pushups", 20, "", "done")
		db.LogActivity("water", 0, "", "done")

		result, err := getStats()
		require.NoError(t, err)
		require.Len(t, result.Today, 2)
		require.Equal(t, 1, result.Streak)
	})
}
