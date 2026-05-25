package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"AgentTestSoulMCP/internal/persistence"
)

func TestSoulEngine_retrieve_yesterdayTimeWindow(t *testing.T) {
	dir := t.TempDir()
	soulPath := filepath.Join(dir, "soul.agent.yaml")
	_ = os.WriteFile(soulPath, []byte("soul: test\n"), 0o644)
	t.Setenv("SOUL_MCP_SOUL_DOC", soulPath)

	agentPath := filepath.Clean(filepath.Join("..", "..", "agentConfig", "soul-agent.yaml"))
	eng, err := NewSoulEngine(dir, "", agentPath)
	if err != nil {
		t.Fatal(err)
	}

	loc := time.Local
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, loc)
	yesterday := time.Date(2026, 5, 21, 9, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	today := now.UTC().Format(time.RFC3339)

	_ = eng.daily.AppendDay([]persistence.Fact{
		{Summary: "昨日讨论 Soul 多通道召回", StoredAt: yesterday, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: yesterday}},
		{Summary: "今日其它", StoredAt: today, Spatiotemporal: persistence.SpatiotemporalTags{Chronos: today}},
	}, time.Date(2026, 5, 21, 0, 0, 0, 0, loc))

	// 通过固定 query + 事实日期验证；Retrieve 内部用 time.Now()，故用「昨天」且事实为真实昨日
	// 集成测：写入相对 calendar 的 yesterday 字符串，query 用 ISO 日期更稳
	yDate := time.Date(2026, 5, 21, 0, 0, 0, 0, loc).Format("2006-01-02")
	ret := eng.Retrieve(context.Background(), RetrieveInput{
		Context: "用户输入: " + yDate + " 讨论了什么\n", QueryHint: yDate + " 讨论了什么",
	})
	var p struct {
		Hints string `json:"hints"`
	}
	if err := json.Unmarshal([]byte(ret), &p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.Hints, "Soul") && !strings.Contains(p.Hints, "多通道") {
		t.Fatalf("hints should include yesterday fact: %q", p.Hints)
	}
}
