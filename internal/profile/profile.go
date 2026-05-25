package profile

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry profile.jsonl 一行。
type Entry struct {
	Key             string  `json:"key"`
	Value           string  `json:"value"`
	Confidence      float64 `json:"confidence"`
	SourceEpisodeID string  `json:"source_episode_id"`
	Evidence        string  `json:"evidence,omitempty"`
	UpdatedAt       string  `json:"updated_at"`
}

// Repo JSONL 存储。
type Repo struct {
	path string
}

func NewRepo(dataDir string) (*Repo, error) {
	if err := os.MkdirAll(filepath.Join(dataDir, "profile"), 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(dataDir, "profile.jsonl")
	return &Repo{path: p}, nil
}

func (r *Repo) List() ([]Entry, error) {
	f, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	byKey := map[string]Entry{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Entry
		if json.Unmarshal([]byte(line), &e) != nil || e.Key == "" {
			continue
		}
		if prev, ok := byKey[e.Key]; !ok || e.UpdatedAt > prev.UpdatedAt {
			byKey[e.Key] = e
		}
	}
	out := make([]Entry, 0, len(byKey))
	for _, e := range byKey {
		out = append(out, e)
	}
	return out, sc.Err()
}

func (r *Repo) Upsert(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	f, err := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	for _, e := range entries {
		if e.UpdatedAt == "" {
			e.UpdatedAt = now
		}
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return nil
}
