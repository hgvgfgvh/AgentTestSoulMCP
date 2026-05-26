package engine

import (
	"path/filepath"
	"testing"
	"time"

	"AgentTestSoulMCP/internal/persistence"
	"AgentTestSoulMCP/internal/soulagent"
)

func TestLoadFactsForRetrieve_lastMonthOnDisk(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Clean(filepath.Join("..", "..", "agentConfig", "soul-agent.yaml"))
	eng, err := NewSoulEngine(dir, "", agentPath)
	if err != nil {
		t.Fatal(err)
	}
	loc := time.Local
	april := time.Date(2026, 4, 15, 10, 0, 0, 0, loc)
	ch := april.UTC().Format(time.RFC3339)
	_ = eng.daily.AppendDay([]persistence.Fact{
		{Summary: "四月讨论 Game01", StoredAt: ch, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: ch}},
	}, april)

	got := eng.loadFactsForRetrieve("上个月讨论了什么", nil)
	found := false
	for _, f := range got {
		if f.Summary == "四月讨论 Game01" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected April fact in dynamic load, got %d facts", len(got))
	}

	tags := &soulagent.RetrievalTags{DateHints: []string{"2026-04"}}
	got2 := eng.loadFactsForRetrieve("讨论汇总", tags)
	found = false
	for _, f := range got2 {
		if f.Summary == "四月讨论 Game01" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("date_hints month should load April file")
	}
}

func TestLoadFactsForRetrieve_onlyLoadsRequestedDays(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Clean(filepath.Join("..", "..", "agentConfig", "soul-agent.yaml"))
	eng, err := NewSoulEngine(dir, "", agentPath)
	if err != nil {
		t.Fatal(err)
	}
	loc := time.Local
	yesterday := time.Date(2026, 5, 21, 9, 0, 0, 0, loc)
	today := time.Date(2026, 5, 22, 9, 0, 0, 0, loc)
	_ = eng.daily.AppendDay([]persistence.Fact{{Summary: "only-yesterday", StoredAt: yesterday.UTC().Format(time.RFC3339)}}, yesterday)
	_ = eng.daily.AppendToday([]persistence.Fact{{Summary: "today-fact", StoredAt: today.UTC().Format(time.RFC3339)}})

	got := eng.loadFactsForRetrieve("2026-05-21 讨论了什么", nil)
	var hasY, hasT bool
	for _, f := range got {
		if f.Summary == "only-yesterday" {
			hasY = true
		}
		if f.Summary == "today-fact" {
			hasT = true
		}
	}
	if !hasY || hasT {
		t.Fatalf("ISO day query should load 2026-05-21 only: hasY=%v hasT=%v len=%d", hasY, hasT, len(got))
	}
}
