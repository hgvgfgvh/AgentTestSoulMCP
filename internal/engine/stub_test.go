package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStubEngine_StoreRetrieve(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "soul.config")
	_ = os.WriteFile(cfgPath, []byte("persona:\n  role: 测试人格\n"), 0o644)

	eng, err := NewStubEngine(dir, cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	storeOut := eng.Store(context.Background(), StoreInput{
		Content:       "[source=agenttest-webui]\n\n## 用户\n聊论文",
		Source:        "agenttest-webui",
		Kind:          "dialogue",
		CorrelationID: "t1",
	})
	if !strings.Contains(storeOut, `"accepted":"true"`) {
		t.Fatalf("store: %s", storeOut)
	}
	retOut := eng.Retrieve(context.Background(), RetrieveInput{Context: "用户输入: 昨天论文"})
	var payload struct {
		Hints string `json:"hints"`
	}
	if err := json.Unmarshal([]byte(retOut), &payload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(payload.Hints, "测试人格") {
		t.Fatalf("hints missing persona: %q", payload.Hints)
	}
}
