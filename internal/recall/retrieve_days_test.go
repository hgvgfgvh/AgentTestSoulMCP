package recall

import (
	"testing"
	"time"

	"AgentTestSoulMCP/internal/soulagent"
)

func TestResolveRetrieveDayKeys_lastMonth(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local)
	keys := ResolveRetrieveDayKeys("上个月讨论了什么", nil, now, 7, 90)
	if len(keys) < 28 || len(keys) > 31 {
		t.Fatalf("last month should be ~28-31 days, got %d: %v", len(keys), keys)
	}
	if keys[0] != "2026-04-01" || keys[len(keys)-1] != "2026-04-30" {
		t.Fatalf("unexpected range: %s .. %s", keys[0], keys[len(keys)-1])
	}
}

func TestResolveRetrieveDayKeys_dateHintMonth(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local)
	tags := &soulagent.RetrievalTags{DateHints: []string{"2026-04"}}
	keys := ResolveRetrieveDayKeys("讨论了什么", tags, now, 7, 90)
	if len(keys) != 30 {
		t.Fatalf("April 2026 has 30 days, got %d", len(keys))
	}
}

func TestResolveRetrieveDayKeys_defaultRecent(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.Local)
	keys := ResolveRetrieveDayKeys("LibGDX 锁", nil, now, 7, 90)
	if len(keys) != 7 {
		t.Fatalf("want 7 default days, got %d", len(keys))
	}
	if keys[0] != "2026-05-22" || keys[6] != "2026-05-16" {
		t.Fatalf("keys=%v", keys)
	}
}

func TestParseCues_lastMonth(t *testing.T) {
	now := time.Date(2026, 5, 22, 15, 0, 0, 0, time.Local)
	c := ParseCues("上个月", now)
	if !c.Time.Active {
		t.Fatal("expected active")
	}
	wantStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	if !c.Time.Start.Equal(wantStart) {
		t.Fatalf("start=%v", c.Time.Start)
	}
}
