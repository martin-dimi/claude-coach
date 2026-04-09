package config

import (
	"testing"
	"time"
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
		{"", 5 * time.Minute},       // fallback
		{"garbage", 5 * time.Minute}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseDuration(tt.input, 5*time.Minute)
			if got != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestActivityIntervalDuration(t *testing.T) {
	a := Activity{Name: "pushups", Interval: "1h"}
	if a.IntervalDuration() != time.Hour {
		t.Errorf("expected 1h, got %v", a.IntervalDuration())
	}

	a = Activity{Name: "water", Interval: "30m"}
	if a.IntervalDuration() != 30*time.Minute {
		t.Errorf("expected 30m, got %v", a.IntervalDuration())
	}
}

func TestSkipCooldownDuration(t *testing.T) {
	s := Settings{SkipCooldown: "10m"}
	if s.SkipCooldownDuration() != 10*time.Minute {
		t.Errorf("expected 10m, got %v", s.SkipCooldownDuration())
	}

	s = Settings{SkipCooldown: ""}
	if s.SkipCooldownDuration() != 10*time.Minute {
		t.Errorf("expected 10m fallback, got %v", s.SkipCooldownDuration())
	}
}

func TestLoadSaveConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// No config yet
	if Exists() {
		t.Fatal("config should not exist yet")
	}

	cfg := DefaultConfig()
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	if !Exists() {
		t.Fatal("config should exist after save")
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Activities) != len(cfg.Activities) {
		t.Fatalf("expected %d activities, got %d", len(cfg.Activities), len(loaded.Activities))
	}

	if loaded.Settings.ActiveHours != cfg.Settings.ActiveHours {
		t.Fatalf("active hours mismatch: %v vs %v", loaded.Settings.ActiveHours, cfg.Settings.ActiveHours)
	}
}
