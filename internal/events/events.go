package events

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"AgentTestSoulMCP/internal/textutil"
)

// Entry events.jsonl 一行。
type Entry struct {
	EventID         string   `json:"event_id"`
	Kind            string   `json:"kind"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary,omitempty"`
	Entities        []string `json:"entities,omitempty"`
	LastMentioned   string   `json:"last_mentioned"`
	EvidenceSnippet string   `json:"evidence_snippet,omitempty"`
	SourceEpisodeID string   `json:"source_episode_id"`
}

// Repo JSONL 存储；读取时按 title 归并保留最新。
type Repo struct {
	path     string
	mergeCos float64
}

func NewRepo(dataDir string, mergeCosine float64) (*Repo, error) {
	if mergeCosine <= 0 {
		mergeCosine = 0.85
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "episodes"), 0o755); err != nil {
		return nil, err
	}
	return &Repo{
		path:     filepath.Join(dataDir, "events.jsonl"),
		mergeCos: mergeCosine,
	}, nil
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
	byTitle := map[string]Entry{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Entry
		if json.Unmarshal([]byte(line), &e) != nil || e.Title == "" {
			continue
		}
		key := normalizeTitle(e.Title)
		if prev, ok := byTitle[key]; !ok || e.LastMentioned > prev.LastMentioned {
			byTitle[key] = e
		}
	}
	out := make([]Entry, 0, len(byTitle))
	for _, e := range byTitle {
		out = append(out, e)
	}
	return out, sc.Err()
}

// Upsert 追加新行；同标题高重合则更新 last_mentioned 与 summary。
func (r *Repo) Upsert(incoming []Entry, episodeID string) error {
	if len(incoming) == 0 {
		return nil
	}
	existing, _ := r.List()
	now := time.Now().UTC().Format(time.RFC3339)
	merged := append([]Entry{}, existing...)
	for _, neu := range incoming {
		if neu.EventID == "" {
			neu.EventID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
		}
		if neu.LastMentioned == "" {
			neu.LastMentioned = now
		}
		if neu.SourceEpisodeID == "" {
			neu.SourceEpisodeID = episodeID
		}
		found := false
		nt := normalizeTitle(neu.Title)
		for i, old := range merged {
			if textutil.OverlapScore(nt, normalizeTitle(old.Title)) >= r.mergeCos {
				merged[i].LastMentioned = neu.LastMentioned
				merged[i].SourceEpisodeID = neu.SourceEpisodeID
				if neu.Summary != "" {
					merged[i].Summary = neu.Summary
				}
				if len(neu.Entities) > 0 {
					merged[i].Entities = mergeStrings(merged[i].Entities, neu.Entities)
				}
				if neu.EvidenceSnippet != "" {
					merged[i].EvidenceSnippet = neu.EvidenceSnippet
				}
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, neu)
		}
	}
	return r.rewrite(merged)
}

func (r *Repo) rewrite(all []Entry) error {
	f, err := os.OpenFile(r.path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, e := range all {
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

func normalizeTitle(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func mergeStrings(a, b []string) []string {
	m := map[string]bool{}
	var out []string
	for _, x := range append(a, b...) {
		x = strings.TrimSpace(x)
		if x == "" || m[x] {
			continue
		}
		m[x] = true
		out = append(out, x)
	}
	return out
}
