package db

import (
	"os"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	conn = nil
	dir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() { os.Unsetenv("XDG_CONFIG_HOME"); conn = nil })
}

func TestLogAndLastTime(t *testing.T) {
	setupTestDB(t)

	// No records yet
	last, err := LastTime("pushups", "done")
	if err != nil {
		t.Fatal(err)
	}
	if !last.IsZero() {
		t.Fatalf("expected zero time, got %v", last)
	}

	// Log an activity
	if err := LogActivity("pushups", 20, "", "done"); err != nil {
		t.Fatal(err)
	}

	last, err = LastTime("pushups", "done")
	if err != nil {
		t.Fatal(err)
	}
	if last.IsZero() {
		t.Fatal("expected non-zero time after logging")
	}
	if time.Since(last) > 5*time.Second {
		t.Fatalf("last time too old: %v", last)
	}

	// Different activity should still be zero
	last, err = LastTime("water", "done")
	if err != nil {
		t.Fatal(err)
	}
	if !last.IsZero() {
		t.Fatalf("expected zero time for water, got %v", last)
	}
}

func TestLastTimeSeparatesActions(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 0, "", "skip")

	lastDone, _ := LastTime("pushups", "done")
	lastSkip, _ := LastTime("pushups", "skip")

	if lastDone.IsZero() || lastSkip.IsZero() {
		t.Fatal("both done and skip should have times")
	}
}

func TestStats(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 0, "", "skip")
	LogActivity("water", 0, "", "done")

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stats, err := Stats(startOfDay, now)
	if err != nil {
		t.Fatal(err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(stats))
	}

	// Stats are ordered by activity name
	pushups := stats[0]
	if pushups.Activity != "pushups" {
		t.Fatalf("expected pushups, got %s", pushups.Activity)
	}
	if pushups.TotalReps != 40 {
		t.Fatalf("expected 40 reps, got %d", pushups.TotalReps)
	}
	if pushups.DoneCount != 2 {
		t.Fatalf("expected 2 done, got %d", pushups.DoneCount)
	}
	if pushups.SkipCount != 1 {
		t.Fatalf("expected 1 skip, got %d", pushups.SkipCount)
	}

	water := stats[1]
	if water.Activity != "water" {
		t.Fatalf("expected water, got %s", water.Activity)
	}
	if water.DoneCount != 1 {
		t.Fatalf("expected 1 done for water, got %d", water.DoneCount)
	}
}

func TestStatsTimeRange(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")

	now := time.Now()

	// Future range should return nothing
	future := now.Add(1 * time.Hour)
	stats, err := Stats(future, future.Add(1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected 0 stats for future range, got %d", len(stats))
	}

	// Lifetime (zero to now) should include it
	stats, err = Stats(time.Time{}, now.Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat for lifetime, got %d", len(stats))
	}
}

func TestStatsByDay(t *testing.T) {
	setupTestDB(t)

	LogActivity("pushups", 20, "", "done")
	LogActivity("pushups", 20, "", "done")
	LogActivity("water", 0, "", "done")

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stats, err := StatsByDay(startOfDay, now.Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 day stats, got %d", len(stats))
	}

	today := now.Format("2006-01-02")
	for _, s := range stats {
		if s.Date != today {
			t.Fatalf("expected date %s, got %s", today, s.Date)
		}
	}
}

func TestCurrentStreak(t *testing.T) {
	setupTestDB(t)

	// No data = no streak
	current, best, err := CurrentStreak()
	if err != nil {
		t.Fatal(err)
	}
	if current != 0 || best != 0 {
		t.Fatalf("expected 0/0 streak, got %d/%d", current, best)
	}

	// Log today
	LogActivity("pushups", 20, "", "done")
	current, _, err = CurrentStreak()
	if err != nil {
		t.Fatal(err)
	}
	if current != 1 {
		t.Fatalf("expected 1 day streak, got %d", current)
	}

	// Insert a record for yesterday manually
	db, _ := Open()
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", yesterday)

	current, best, err = CurrentStreak()
	if err != nil {
		t.Fatal(err)
	}
	if current != 2 {
		t.Fatalf("expected 2 day streak, got %d", current)
	}
	if best < 2 {
		t.Fatalf("expected best >= 2, got %d", best)
	}
}

func TestCurrentStreakGap(t *testing.T) {
	setupTestDB(t)

	db, _ := Open()

	// Insert records with a gap: today, yesterday, 3 days ago (missing 2 days ago)
	today := time.Now().Format("2006-01-02 15:04:05")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02 15:04:05")

	db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", today)
	db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", yesterday)
	db.Exec("INSERT INTO activity_log (activity, reps, action, created_at) VALUES ('pushups', 20, 'done', ?)", threeDaysAgo)

	current, best, _ := CurrentStreak()
	if current != 2 {
		t.Fatalf("expected current streak 2 (gap at day -2), got %d", current)
	}
	if best != 2 {
		t.Fatalf("expected best streak 2, got %d", best)
	}
}
