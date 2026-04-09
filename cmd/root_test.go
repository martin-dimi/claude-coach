package cmd

import (
	"testing"

	"github.com/fridge/coach/internal/config"
	"github.com/stretchr/testify/require"
)

func TestWithinActiveHours(t *testing.T) {
	tests := []struct {
		name   string
		hours  [2]string
		expect bool
	}{
		{"all day", [2]string{"00:00", "23:59"}, true},
		{"invalid format defaults to true", [2]string{"bad", "format"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := config.Settings{ActiveHours: tt.hours}
			require.Equal(t, tt.expect, withinActiveHours(s))
		})
	}
}

func TestFindActivity(t *testing.T) {
	cfg := testConfig(
		config.Activity{Name: "pushups", Reps: 20},
		config.Activity{Name: "water", Message: "Drink"},
	)

	tests := []struct {
		name   string
		search string
		found  bool
		reps   int
	}{
		{"existing", "pushups", true, 20},
		{"another", "water", true, 0},
		{"missing", "squats", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, found := findActivity(cfg, tt.search)
			require.Equal(t, tt.found, found)
			if found {
				require.Equal(t, tt.reps, a.Reps)
			}
		})
	}
}
