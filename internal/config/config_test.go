package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"10m", 10 * time.Minute},
		{"1h", 1 * time.Hour},
		{"30s", 30 * time.Second},
		{"2h", 2 * time.Hour},
		{"1h30m", 90 * time.Minute},
		{"", 5 * time.Minute},
		{"garbage", 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, ParseDuration(tt.input, 5*time.Minute))
		})
	}
}

func TestActivityIntervalDuration(t *testing.T) {
	tests := []struct {
		interval string
		expected time.Duration
	}{
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			a := Activity{Name: "test", Interval: tt.interval}
			require.Equal(t, tt.expected, a.IntervalDuration())
		})
	}
}

func TestSkipCooldownDuration(t *testing.T) {
	tests := []struct {
		name     string
		cooldown string
		expected time.Duration
	}{
		{"configured", "10m", 10 * time.Minute},
		{"empty fallback", "", 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Settings{SkipCooldown: tt.cooldown}
			require.Equal(t, tt.expected, s.SkipCooldownDuration())
		})
	}
}

func TestLoadSaveConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	require.False(t, Exists())

	cfg := DefaultConfig()
	require.NoError(t, Save(cfg))
	require.True(t, Exists())

	loaded, err := Load()
	require.NoError(t, err)
	require.Equal(t, len(cfg.Activities), len(loaded.Activities))
	require.Equal(t, cfg.Settings.ActiveHours, loaded.Settings.ActiveHours)
}
