package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/filter"
	"AgentTestSoulMCP/internal/persistence"
	"AgentTestSoulMCP/internal/response"
)

// SoulEngine v4：按天落地 + person + map + llm_cache；异步四路 store；快慢双轨 retrieve。
type SoulEngine struct {
	dataDir  string
	agentCfg *config.AgentConfig
	paths    config.DataPaths
	daily    *persistence.DailyHistoryStore
	person   *persistence.PersonDoc
	mapDoc   *persistence.MapDoc
	cache    *persistence.LLMCacheStore
	soul     *persistence.SoulDoc
	legacy   *persistence.HistoryStore
	jobSeq   atomic.Uint64
	wg       sync.WaitGroup
}

// NewSoulEngine 构造引擎。
func NewSoulEngine(dataDir, _ /*legacy soul config*/, agentConfigPath string) (*SoulEngine, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	ac, err := config.LoadAgentConfig(agentConfigPath)
	if err != nil {
		return nil, err
	}
	paths := ac.ResolveDataPaths(dataDir)
	ensurePersonTemplate(paths.Person)
	ensureMapTemplate(paths.Map)

	var legacy *persistence.HistoryStore
	if paths.LegacyHistory != "" {
		legacy = persistence.NewHistoryStore(paths.LegacyHistory)
	}

	return &SoulEngine{
		dataDir:  dataDir,
		agentCfg: ac,
		paths:    paths,
		daily:    persistence.NewDailyHistoryStore(paths.HistoryDir),
		person:   persistence.NewPersonDoc(paths.Person),
		mapDoc:   persistence.NewMapDoc(paths.Map),
		cache:    persistence.NewLLMCacheStore(paths.LLMCache),
		soul:     persistence.NewSoulDoc(paths.Soul),
		legacy:   legacy,
	}, nil
}

func (e *SoulEngine) Store(ctx context.Context, in StoreInput) string {
	_ = ctx
	if skip, reason := filter.ShouldSkipStore(in.Content); skip && e.agentCfg.Store.SkipChitchat {
		return response.FormatStore(response.StorePayload{
			Accepted: "false", Skipped: "true", SkipReason: reason,
			Message: "soul store skipped", Phase: response.PhaseSoul(),
		})
	}
	jobID := formatJobID(e.jobSeq.Add(1))
	if strings.TrimSpace(in.CorrelationID) != "" {
		jobID = "soul-" + strings.TrimSpace(in.CorrelationID)
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.processStore(jobID, in)
	}()
	return response.FormatStore(response.StorePayload{
		Accepted: "true", JobID: jobID, Skipped: "false",
		Message: "accepted; async store (daily+person+map+prefetch)",
		Phase:   response.PhaseSoul(),
	})
}

func (e *SoulEngine) Retrieve(ctx context.Context, in RetrieveInput) string {
	if skip, reason := filter.ShouldSkipRetrieve(in.Context); skip {
		return response.FormatRetrieve(response.RetrievePayload{
			Hints: "", Skipped: "true", SkipReason: reason, Phase: response.PhaseSoul(),
		})
	}
	query := filter.ExtractUserQuery(in.Context)
	if strings.TrimSpace(in.QueryHint) != "" {
		query = strings.TrimSpace(in.QueryHint)
	}
	hints := e.runRetrieve(ctx, query)
	return response.FormatRetrieve(response.RetrievePayload{
		Hints: hints, Skipped: "false", Phase: response.PhaseSoul(),
	})
}

func formatJobID(n uint64) string {
	return fmt.Sprintf("soul-job-%d-%d", time.Now().Unix(), n)
}

func ensurePersonTemplate(personPath string) {
	if _, err := os.Stat(personPath); err == nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(personPath), 0o755)
	_ = os.WriteFile(personPath, []byte("# 用户画像\n\n（待 Soul store 归纳）\n"), 0o644)
}

func ensureMapTemplate(mapPath string) {
	if _, err := os.Stat(mapPath); err == nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(mapPath), 0o755)
	_ = os.WriteFile(mapPath, []byte(persistenceMapTemplate()), 0o644)
}

func persistenceMapTemplate() string {
	return `# Soul 地图（热索引）

> 高价值摘要 + 落地文件指针。

## 条目

（尚无记录）
`
}
