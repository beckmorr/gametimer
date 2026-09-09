package session

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/beckmorr/gametimer/internal/config"
)

type Status string

const (
	StatusCompleted   Status = "completed"
	StatusSavedInTime Status = "saved_in_time"
	StatusCancelled   Status = "cancelled"
)

type Entry struct {
	Game       string    `json:"game"`
	Process    string    `json:"process,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	PlannedMin int       `json:"planned_min"`
	ExtraMin   int       `json:"extra_min"`
	PlayedMin  float64   `json:"played_min"`
	Status     Status    `json:"status"`
}

func Append(e Entry) error {
	path, err := config.LogPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func Load() ([]Entry, error) {
	path, err := config.LogPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// Remove deletes entries matching gameFilter (if non-empty) and ended at or
// after since (if non-zero), rewriting the log with the rest. It returns how
// many entries were removed.
func Remove(gameFilter string, since time.Time) (int, error) {
	path, err := config.LogPath()
	if err != nil {
		return 0, err
	}
	entries, err := Load()
	if err != nil {
		return 0, err
	}

	kept := entries[:0]
	removed := 0
	for _, e := range entries {
		match := (gameFilter == "" || e.Game == gameFilter) && (since.IsZero() || !e.EndedAt.Before(since))
		if match {
			removed++
			continue
		}
		kept = append(kept, e)
	}
	if removed == 0 {
		return 0, nil
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	for _, e := range kept {
		data, err := json.Marshal(e)
		if err != nil {
			return 0, err
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			return 0, err
		}
	}
	return removed, nil
}

type Stats struct {
	TotalSessions int
	TotalMinutes  float64
	ByGame        map[string]float64
}

func Aggregate(entries []Entry, gameFilter string, since time.Time) Stats {
	stats := Stats{ByGame: map[string]float64{}}
	for _, e := range entries {
		if gameFilter != "" && e.Game != gameFilter {
			continue
		}
		if !since.IsZero() && e.EndedAt.Before(since) {
			continue
		}
		stats.TotalSessions++
		stats.TotalMinutes += e.PlayedMin
		stats.ByGame[e.Game] += e.PlayedMin
	}
	return stats
}
