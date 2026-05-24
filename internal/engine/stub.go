package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"AgentTestSoulMCP/internal/response"

	"gopkg.in/yaml.v3"
)

// StubEngine Phase-0：同步 ACK + 异步追加 dialogue 日志；retrieve 返回基座占位 hints。
type StubEngine struct {
	dataDir    string
	configPath string
	jobSeq     atomic.Uint64
	wg         sync.WaitGroup
}

// NewStubEngine dataDir 为空时使用 ./data；configPath 为 soul.config 路径（可空）。
func NewStubEngine(dataDir, configPath string) (*StubEngine, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	return &StubEngine{dataDir: dataDir, configPath: strings.TrimSpace(configPath)}, nil
}

func (e *StubEngine) Store(ctx context.Context, in StoreInput) string {
	content := strings.TrimSpace(in.Content)
	if len([]rune(content)) < 8 {
		return response.FormatStore(response.StorePayload{
			Accepted:   "false",
			Skipped:    "true",
			SkipReason: "content_too_short",
			Message:    "soul store skipped",
			Phase:      response.PhaseStub(),
		})
	}
	jobID := fmt.Sprintf("soul-job-%d-%d", time.Now().Unix(), e.jobSeq.Add(1))
	out := response.FormatStore(response.StorePayload{
		Accepted: "true",
		JobID:    jobID,
		Skipped:  "false",
		Message:  "accepted; async stub persist (no profile/events agent yet)",
		Phase:    response.PhaseStub(),
	})
	rec := stubDialogueRecord{
		JobID:         jobID,
		StoredAt:      time.Now().UTC().Format(time.RFC3339),
		Source:        in.Source,
		Kind:          in.Kind,
		CorrelationID: in.CorrelationID,
		ContentLen:    len([]rune(content)),
		Preview:       preview(content, 320),
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		_ = e.appendDialogue(context.Background(), rec)
	}()
	_ = ctx
	return out
}

func (e *StubEngine) Retrieve(ctx context.Context, in RetrieveInput) string {
	_ = ctx
	ctxStr := strings.TrimSpace(in.Context)
	if ctxStr == "" {
		return response.FormatRetrieve(response.RetrievePayload{
			Hints:      "",
			Skipped:    "true",
			SkipReason: "empty_context",
			Phase:      response.PhaseStub(),
		})
	}
	hints := e.buildStubHints(ctxStr)
	return response.FormatRetrieve(response.RetrievePayload{
		Hints:   hints,
		Skipped: "false",
		Phase:   response.PhaseStub(),
	})
}

func (e *StubEngine) buildStubHints(contextStr string) string {
	var b strings.Builder
	b.WriteString("## 协作人格（Soul · stub）\n")
	if persona := loadSoulConfigPersona(e.configPath); persona != "" {
		b.WriteString(persona)
		b.WriteString("\n")
	} else {
		b.WriteString("（soul.config 未加载；核心 Agent 未实现）\n")
	}
	n, last := e.dialogueStats()
	b.WriteString("\n## 近期议题与事件（Soul · stub）\n")
	if n == 0 {
		b.WriteString("（尚无已归档 WebUI 对话；store 后将在此出现摘要）\n")
	} else {
		b.WriteString(fmt.Sprintf("已归档 WebUI 对话 %d 条；最近: %s\n", n, last))
	}
	b.WriteString("\n---\n检索上下文摘录:\n")
	b.WriteString(preview(contextStr, 400))
	return b.String()
}

type stubDialogueRecord struct {
	JobID         string `json:"job_id"`
	StoredAt      string `json:"stored_at"`
	Source        string `json:"source"`
	Kind          string `json:"kind"`
	CorrelationID string `json:"correlation_id"`
	ContentLen    int    `json:"content_len"`
	Preview       string `json:"preview"`
}

func (e *StubEngine) appendDialogue(ctx context.Context, rec stubDialogueRecord) error {
	_ = ctx
	path := filepath.Join(e.dataDir, "dialogues.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func (e *StubEngine) dialogueStats() (count int, lastAt string) {
	path := filepath.Join(e.dataDir, "dialogues.jsonl")
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return 0, ""
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	count = len(lines)
	var rec stubDialogueRecord
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); err == nil {
		lastAt = rec.StoredAt
	}
	return count, lastAt
}

func loadSoulConfigPersona(configPath string) string {
	if configPath == "" {
		return ""
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	var doc struct {
		Persona struct {
			Role string `yaml:"role"`
		} `yaml:"persona"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return ""
	}
	return strings.TrimSpace(doc.Persona.Role)
}

func preview(s string, maxRunes int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= maxRunes {
		return string(r)
	}
	return string(r[:maxRunes]) + "…"
}
