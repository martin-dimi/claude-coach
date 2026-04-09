package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Activity struct {
	Name     string `toml:"name"`
	Reps     int    `toml:"reps,omitempty"`
	Duration string `toml:"duration,omitempty"`
	Message  string `toml:"message,omitempty"`
	Interval string `toml:"interval"`
}

func (a Activity) IntervalDuration() time.Duration {
	return ParseDuration(a.Interval, 60*time.Minute)
}

type Settings struct {
	ActiveHours  [2]string `toml:"active_hours"`
	SkipCooldown string    `toml:"skip_cooldown"`
}

func (s Settings) SkipCooldownDuration() time.Duration {
	return ParseDuration(s.SkipCooldown, 10*time.Minute)
}

type Config struct {
	Settings   Settings   `toml:"settings"`
	Activities []Activity `toml:"activities"`
}

func DefaultConfig() Config {
	return Config{
		Settings: Settings{
			ActiveHours:  [2]string{"09:00", "18:00"},
			SkipCooldown: "10m",
		},
		Activities: []Activity{
			{Name: "pushups", Reps: 20, Interval: "1h"},
			{Name: "water", Message: "Drink a glass of water", Interval: "30m"},
			{Name: "stretch", Duration: "2m", Interval: "2h"},
		},
	}
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "coach")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "coach")
}

func Path() string {
	return filepath.Join(Dir(), "config.toml")
}

func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}

func Load() (Config, error) {
	var cfg Config
	path := Path()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, os.ErrNotExist
	}

	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}

func Save(cfg Config) error {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return err
	}
	f, err := os.Create(Path())
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func ParseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	// Handle simple suffixes: 10m, 1h, 2h, 30s
	if len(s) > 1 {
		numStr := s[:len(s)-1]
		suffix := s[len(s)-1]
		if n, err := strconv.Atoi(numStr); err == nil {
			switch suffix {
			case 'm':
				return time.Duration(n) * time.Minute
			case 'h':
				return time.Duration(n) * time.Hour
			case 's':
				return time.Duration(n) * time.Second
			}
		}
	}
	// Handle "1h30m" style via strings check
	if strings.ContainsAny(s, "hms") {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return fallback
}
