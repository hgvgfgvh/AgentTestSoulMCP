package recall

import (
	"testing"
	"time"

	"AgentTestSoulMCP/internal/persistence"
)

func TestParseCues_yesterday(t *testing.T) {
	now := time.Date(2026, 5, 22, 15, 0, 0, 0, time.Local)
	c := ParseCues("昨天我问了什么", now)
	if !c.Time.Active {
		t.Fatal("expected time window")
	}
	wantStart := time.Date(2026, 5, 21, 0, 0, 0, 0, time.Local)
	if !c.Time.Start.Equal(wantStart) {
		t.Fatalf("start=%v want %v", c.Time.Start, wantStart)
	}
}

func TestSelect_yesterdayWithoutKeywordOverlap(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, loc)
	yesterday := time.Date(2026, 5, 21, 10, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	today := time.Date(2026, 5, 22, 9, 0, 0, 0, loc).UTC().Format(time.RFC3339)

	facts := []persistence.Fact{
		{ID: "old", Summary: "完全无关的闲聊", StoredAt: yesterday, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: yesterday}},
		{ID: "today", Summary: "今天的计划", StoredAt: today, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: today}},
	}
	got := Select(facts, "昨天我问了什么", 5, now)
	if len(got) == 0 || got[0].ID != "old" {
		t.Fatalf("got=%+v", got)
	}
}

func TestSelect_entityAndTextFusion(t *testing.T) {
	now := time.Now()
	facts := []persistence.Fact{
		{Summary: "天气不错", Phenomenon: persistence.PhenomenonTags{Category: []string{"Topic"}}},
		{Summary: "修复 LibGDX 锁", Phenomenon: persistence.PhenomenonTags{
			Entity: []string{"LibGDX-RenderPipeline"}, Category: []string{"CodeBug"},
		}, StoredAt: now.UTC().Format(time.RFC3339), Spatiotemporal: persistence.SpatiotemporalTags{Chronos: now.UTC().Format(time.RFC3339)}},
	}
	got := Select(facts, "LibGDX CodeBug", 5, now)
	if len(got) == 0 || got[0].Summary != "修复 LibGDX 锁" {
		t.Fatalf("got=%+v", got)
	}
}

func TestSelect_pitfallOutcome(t *testing.T) {
	now := time.Now()
	ch := now.UTC().Format(time.RFC3339)
	facts := []persistence.Fact{
		{Summary: "成功部署", Causality: persistence.CausalityTags{Outcome: "Success_Route"}, StoredAt: ch, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: ch}},
		{Summary: "安装依赖失败", Causality: persistence.CausalityTags{Outcome: "Pitfall_Route"}, StoredAt: ch, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: ch}},
	}
	got := Select(facts, "上次失败怎么回事", 5, now)
	if len(got) == 0 || got[0].Summary != "安装依赖失败" {
		t.Fatalf("got=%+v", got)
	}
}

func TestFuseRRF(t *testing.T) {
	r1 := []int{0, 1}
	r2 := []int{1, 0}
	out := fuseRRF([][]int{r1, r2}, 2, 60)
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
}
