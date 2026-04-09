package config

import (
	"time"
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
