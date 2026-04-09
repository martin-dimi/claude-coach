package cmd

import (
	"testing"

	"github.com/martin-dimi/claude-coach/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLogSkip(t *testing.T) {
	tests := []struct {
		name             string
		cooldown         string
		expectedCooldown string
	}{
		{"configured cooldown", "15m", "15m"},
		{"default cooldown", "", "10m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			writeTestConfig(t, config.Config{
				Settings: config.Settings{SkipCooldown: tt.cooldown},
			})

			result, err := logSkip("pushups")
			require.NoError(t, err)
			require.True(t, result.Skipped)
			require.Equal(t, tt.expectedCooldown, result.Cooldown)
		})
	}
}
