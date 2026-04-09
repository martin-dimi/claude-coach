package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var conn *sql.DB

func dbPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "coach", "coach.db")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "coach", "coach.db")
}

// Reset closes the connection so the next Open() creates a fresh one.
// Used by tests.
func Reset() {
	if conn != nil {
		conn.Close()
		conn = nil
	}
}

func Open() (*sql.DB, error) {
	if conn != nil {
		return conn, nil
	}

	dir := filepath.Dir(dbPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	var err error
	conn, err = sql.Open("sqlite", dbPath())
	if err != nil {
		return nil, err
	}

	return conn, migrate(conn)
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS activity_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity TEXT NOT NULL,
			reps INTEGER NOT NULL DEFAULT 0,
			duration TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

func LogActivity(activity string, reps int, duration string, action string) error {
	db, err := Open()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"INSERT INTO activity_log (activity, reps, duration, action) VALUES (?, ?, ?, ?)",
		activity, reps, duration, action,
	)
	return err
}

func LastTime(activity, action string) (time.Time, error) {
	db, err := Open()
	if err != nil {
		return time.Time{}, err
	}
	var ts string
	err = db.QueryRow(
		"SELECT created_at FROM activity_log WHERE activity = ? AND action = ? ORDER BY created_at DESC LIMIT 1",
		activity, action,
	).Scan(&ts)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", ts)
}

type ActivityStat struct {
	Activity  string `json:"activity"`
	TotalReps int    `json:"total_reps"`
	DoneCount int    `json:"done_count"`
	SkipCount int    `json:"skip_count"`
}

func Stats(from, to time.Time) ([]ActivityStat, error) {
	db, err := Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT activity,
			COALESCE(SUM(CASE WHEN action='done' THEN reps ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN action='done' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN action='skip' THEN 1 ELSE 0 END), 0)
		FROM activity_log
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY activity
		ORDER BY activity
	`, from.UTC().Format("2006-01-02 15:04:05"), to.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ActivityStat
	for rows.Next() {
		var s ActivityStat
		if err := rows.Scan(&s.Activity, &s.TotalReps, &s.DoneCount, &s.SkipCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

type DayStat struct {
	Date     string
	Activity string
	Reps     int
	Done     int
}

func StatsByDay(from, to time.Time) ([]DayStat, error) {
	db, err := Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT date(created_at) as day, activity,
			COALESCE(SUM(CASE WHEN action='done' THEN reps ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN action='done' THEN 1 ELSE 0 END), 0)
		FROM activity_log
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY day, activity
		ORDER BY day, activity
	`, from.UTC().Format("2006-01-02 15:04:05"), to.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DayStat
	for rows.Next() {
		var s DayStat
		if err := rows.Scan(&s.Date, &s.Activity, &s.Reps, &s.Done); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func CurrentStreak() (current int, best int, err error) {
	db, err := Open()
	if err != nil {
		return 0, 0, err
	}
	rows, err := db.Query(`
		SELECT DISTINCT date(created_at) as day
		FROM activity_log
		WHERE action = 'done'
		ORDER BY day DESC
	`)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return 0, 0, err
		}
		dates = append(dates, d)
	}

	if len(dates) == 0 {
		return 0, 0, nil
	}

	// Current streak: walk backwards from today
	today := time.Now().Format("2006-01-02")
	expected := today
	for _, d := range dates {
		if d == expected {
			current++
			t, _ := time.Parse("2006-01-02", expected)
			expected = t.AddDate(0, 0, -1).Format("2006-01-02")
		} else if d < expected {
			break
		}
	}

	// Best streak
	run := 1
	for i := 1; i < len(dates); i++ {
		prev, _ := time.Parse("2006-01-02", dates[i-1])
		curr, _ := time.Parse("2006-01-02", dates[i])
		if prev.Sub(curr).Hours() == 24 {
			run++
		} else {
			if run > best {
				best = run
			}
			run = 1
		}
	}
	if run > best {
		best = run
	}

	return current, best, nil
}
