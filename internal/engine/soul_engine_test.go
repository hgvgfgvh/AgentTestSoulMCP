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

func TestSoulEngine_fallbackRetrieve(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "person.md"), []byte("# 用户画像\n\n称呼: 老王\n"), 0o644)
	soulPath := filepath.Join(dir, "soul.agent.yaml")
	_ = os.WriteFile(soulPath, []byte("soul: |\n  测试灵魂\n"), 0o644)
	t.Setenv("SOUL_MCP_SOUL_DOC", soulPath)
	agentPath := filepath.Clean(filepath.Join("..", "..", "agentConfig", "soul-agent.yaml"))
	eng, err := NewSoulEngine(dir, "", agentPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = eng.daily.AppendToday([]persistence.Fact{
		{
			Summary:  "Soul对接测试论文",
			Evidence: "讨论论文",
			Phenomenon: persistence.PhenomenonTags{
				Entity:   []string{"架构论文"},
				Category: []string{"Architecture"},
			},
		},
	})
	// 等待无
	_ = time.Millisecond
	ret := eng.Retrieve(context.Background(), RetrieveInput{
		Context: "用户输入: Soul对接测试论文\n", QueryHint: "Soul对接测试论文",
	})
	var p struct {
		Hints string `json:"hints"`
		Phase string `json:"phase"`
	}
	if err := json.Unmarshal([]byte(ret), &p); err != nil {
		t.Fatal(err)
	}
	if p.Phase != "4-async-pipeline" {
		t.Fatalf("phase=%q", p.Phase)
	}
	if !strings.Contains(p.Hints, "老王") || !strings.Contains(p.Hints, "测试灵魂") {
		t.Fatalf("hints=%q", p.Hints)
	}
}

func TestSoulEngine_fallbackStore(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Clean(filepath.Join("..", "..", "agentConfig", "soul-agent.yaml"))
	eng, err := NewSoulEngine(dir, "", agentPath)
	if err != nil {
		t.Fatal(err)
	}
	out := eng.Store(context.Background(), StoreInput{
		Content: "[source=agenttest-webui]\n\n## 用户（WebUI）\n讨论项目 Alpha 与论文。\n\n## 助手（WebUI）\n好的。",
		Source:  "agenttest-webui",
	})
	if !strings.Contains(out, `"accepted":"true"`) {
		t.Fatal(out)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		facts, _ := eng.daily.ListRecentDays(7, time.Now())
		if len(facts) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("history not written")
}
