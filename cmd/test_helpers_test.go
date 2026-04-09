package cmd

import (
	"os"
	"testing"

	"github.com/fridge/coach/internal/config"
	"github.com/fridge/coach/internal/db"
)

func setupTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() { os.Unsetenv("XDG_CONFIG_HOME") })
	db.Reset()
}

func writeTestConfig(t *testing.T, cfg config.Config) {
	t.Helper()
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
}

func testConfig(activities ...config.Activity) config.Config {
	return config.Config{
		Settings:   config.Settings{ActiveHours: [2]string{"00:00", "23:59"}, SkipCooldown: "10m"},
		Activities: activities,
	}
}
