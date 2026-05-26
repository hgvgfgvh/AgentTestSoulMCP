package engine

import (
	"time"

	"AgentTestSoulMCP/internal/persistence"
	"AgentTestSoulMCP/internal/recall"
	"AgentTestSoulMCP/internal/soulagent"
)

// loadFactsForRetrieve 按 query / Gate 标签动态加载落地日 jsonl（非固定 7 天）。
func (e *SoulEngine) loadFactsForRetrieve(query string, tags *soulagent.RetrievalTags) []persistence.Fact {
	now := time.Now()
	keys := recall.ResolveRetrieveDayKeys(
		query,
		tags,
		now,
		e.agentCfg.Retrieve.DefaultRecentDays,
		e.agentCfg.Retrieve.MaxLoadDays,
	)
	return e.loadFactsByDayKeys(keys)
}

func (e *SoulEngine) loadFactsByDayKeys(dayKeys []string) []persistence.Fact {
	var all []persistence.Fact
	for _, key := range dayKeys {
		facts, err := e.daily.ListDay(key)
		if err != nil {
			continue
		}
		all = append(all, facts...)
	}
	if e.legacy != nil {
		if leg, err := e.legacy.List(); err == nil {
			all = append(all, leg...)
		}
	}
	return all
}

// loadAllFacts 供 store 预取等无用户时间线索的场景：最近 default 天。
func (e *SoulEngine) loadAllFacts() []persistence.Fact {
	return e.loadFactsForRetrieve("", nil)
}
