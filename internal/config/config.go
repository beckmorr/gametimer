package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	appDirName   = "gametimer"
	configFile   = "config.json"
	gamesFile    = "games.json"
	defaultTheme = "default"
	defaultSess  = 90
	defaultExtra = 15
)

type Config struct {
	DefaultTheme      string `json:"default_theme"`
	DefaultSessionMin int    `json:"default_session_min"`
	DefaultExtraMin   int    `json:"default_extra_min"`
}

type Game struct {
	Name       string `json:"name"`
	Process    string `json:"process,omitempty"`
	SessionMin int    `json:"session_min"`
	ExtraMin   int    `json:"extra_min"`
}

func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDirName), nil
}

func DataDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, appDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func ensureConfigDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func LogPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions.jsonl"), nil
}

func Load() (Config, error) {
	cfg := Config{DefaultTheme: defaultTheme, DefaultSessionMin: defaultSess, DefaultExtraMin: defaultExtra}
	dir, err := configDir()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := ensureConfigDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, configFile), data, 0o644)
}

func LoadGames() ([]Game, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, gamesFile))
	if errors.Is(err, os.ErrNotExist) {
		return []Game{}, nil
	}
	if err != nil {
		return nil, err
	}
	var games []Game
	if err := json.Unmarshal(data, &games); err != nil {
		return nil, err
	}
	return games, nil
}

func SaveGames(games []Game) error {
	dir, err := ensureConfigDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(games, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, gamesFile), data, 0o644)
}

func FindGame(games []Game, name string) (Game, bool) {
	for _, g := range games {
		if g.Name == name {
			return g, true
		}
	}
	return Game{}, false
}

func UpsertGame(g Game) ([]Game, error) {
	games, err := LoadGames()
	if err != nil {
		return nil, err
	}
	replaced := false
	for i, existing := range games {
		if existing.Name == g.Name {
			games[i] = g
			replaced = true
			break
		}
	}
	if !replaced {
		games = append(games, g)
	}
	if err := SaveGames(games); err != nil {
		return nil, err
	}
	return games, nil
}

func RemoveGame(name string) (bool, error) {
	games, err := LoadGames()
	if err != nil {
		return false, err
	}
	out := games[:0]
	found := false
	for _, g := range games {
		if g.Name == name {
			found = true
			continue
		}
		out = append(out, g)
	}
	if !found {
		return false, nil
	}
	if err := SaveGames(out); err != nil {
		return false, err
	}
	return true, nil
}
