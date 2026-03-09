package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pager/pagerduty"
)

// Data holds the cached PagerDuty data.
type Data struct {
	Schedules []pagerduty.Schedule `json:"schedules"`
	Users     []pagerduty.User     `json:"users"`
	FetchedAt time.Time            `json:"fetched_at"`
}

// Preferences holds user preferences that persist until cache is cleared.
type Preferences struct {
	LastScheduleID string `json:"last_schedule_id,omitempty"`
}

// Dir returns the cache directory path (~/.cache/pager), creating it if needed.
func Dir() (string, error) {
	return cacheDir()
}

// cacheDir returns ~/.cache/pager, creating it if needed.
func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	dir := filepath.Join(home, ".cache", "pager")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating cache directory: %w", err)
	}
	return dir, nil
}

func cachePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "data.json"), nil
}

func preferencesPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "preferences.json"), nil
}

// Load reads the cache file. Returns nil if cache doesn't exist or can't be read.
func Load() *Data {
	path, err := cachePath()
	if err != nil {
		return nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var data Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}

	return &data
}

// Save writes schedules and users to the cache file.
func Save(schedules []pagerduty.Schedule, users []pagerduty.User) error {
	path, err := cachePath()
	if err != nil {
		return err
	}

	data := Data{
		Schedules: schedules,
		Users:     users,
		FetchedAt: time.Now(),
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache data: %w", err)
	}

	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// LoadLastScheduleID returns the last selected schedule ID, or empty string if none.
func LoadLastScheduleID() string {
	path, err := preferencesPath()
	if err != nil {
		return ""
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var prefs Preferences
	if err := json.Unmarshal(raw, &prefs); err != nil {
		return ""
	}

	return prefs.LastScheduleID
}

// SaveLastScheduleID persists the last selected schedule ID to the preferences file.
func SaveLastScheduleID(id string) error {
	path, err := preferencesPath()
	if err != nil {
		return err
	}

	prefs := Preferences{LastScheduleID: id}

	raw, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling preferences: %w", err)
	}

	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("writing preferences file: %w", err)
	}

	return nil
}

// Clear deletes the cache and preferences files.
func Clear() error {
	dataPath, err := cachePath()
	if err != nil {
		return err
	}

	if err := os.Remove(dataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing cache file: %w", err)
	}

	prefsPath, err := preferencesPath()
	if err != nil {
		return err
	}

	if err := os.Remove(prefsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing preferences file: %w", err)
	}

	return nil
}
