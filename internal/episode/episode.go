package episode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Record 归档到 data/episodes/{jobID}.json。
type Record struct {
	JobID         string `json:"job_id"`
	StoredAt      string `json:"stored_at"`
	Source        string `json:"source"`
	Kind          string `json:"kind"`
	CorrelationID string `json:"correlation_id"`
	Content       string `json:"content"`
}

// Save 写入 episode 文件。
func Save(dataDir, jobID string, rec Record) error {
	dir := filepath.Join(dataDir, "episodes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if rec.StoredAt == "" {
		rec.StoredAt = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, jobID+".json"), b, 0o644)
}
