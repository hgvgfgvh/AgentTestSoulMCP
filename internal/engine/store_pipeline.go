package engine

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"AgentTestSoulMCP/internal/llm"
	"AgentTestSoulMCP/internal/persistence"
	"AgentTestSoulMCP/internal/recall"
	"AgentTestSoulMCP/internal/soulagent"
)

func (e *SoulEngine) processStore(jobID string, in StoreInput) {
	ctx := context.Background()
	client, hasLLM := llm.ConfigFromEnv()
	dialogue := in.Content
	source := strings.TrimSpace(in.Source)
	dayKey := time.Now().Format("2006-01-02")

	personMD, _ := e.person.Read()
	mapMD, _ := e.mapDoc.Read()

	if !hasLLM {
		facts := soulagent.FallbackStore(dialogue, source)
		_ = e.daily.AppendToday(facts)
		return
	}

	var wg sync.WaitGroup
	var dailyFacts []persistence.Fact
	var newPerson, newMap string
	var dailyErr, personErr, mapErr error

	wg.Add(3)
	go func() {
		defer wg.Done()
		dailyFacts, dailyErr = soulagent.RunStoreDailyLLM(ctx, dialogue, dayKey, e.agentCfg, client)
		if dailyErr != nil {
			log.Printf("[soul-mcp] store daily LLM: %v", dailyErr)
			return
		}
		for i := range dailyFacts {
			if dailyFacts[i].Source == "" {
				dailyFacts[i].Source = source
			}
		}
		if err := e.daily.AppendToday(dailyFacts); err != nil {
			log.Printf("[soul-mcp] daily append: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		newPerson, personErr = soulagent.RunStorePersonLLM(ctx, dialogue, personMD, e.agentCfg, client)
		if personErr != nil {
			log.Printf("[soul-mcp] store person LLM: %v", personErr)
			return
		}
		if newPerson != "" {
			if err := e.person.Write(newPerson); err != nil {
				log.Printf("[soul-mcp] person write: %v", err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		recentSummary := e.summarizeRecentDays(e.agentCfg.Store.MapRecentDays)
		newMap, mapErr = soulagent.RunStoreMapLLM(ctx, dialogue, mapMD, recentSummary, e.agentCfg, client)
		if mapErr != nil {
			log.Printf("[soul-mcp] store map LLM: %v", mapErr)
			return
		}
		if newMap != "" {
			if err := e.mapDoc.Write(newMap); err != nil {
				log.Printf("[soul-mcp] map write: %v", err)
			}
		}
	}()
	wg.Wait()

	// 任务4：预取缓存（依赖已更新的 person/map/daily）
	personMD, _ = e.person.Read()
	mapMD, _ = e.mapDoc.Read()
	e.buildPrefetchCache(ctx, jobID, dialogue, personMD, mapMD, client)
}

func (e *SoulEngine) summarizeRecentDays(days int) string {
	facts, err := e.daily.ListRecentDays(days, time.Now())
	if err != nil || len(facts) == 0 {
		return "（无最近落地记录）"
	}
	var lines []string
	for i, f := range facts {
		if i >= 20 {
			break
		}
		lines = append(lines, "- "+f.Summary)
	}
	return strings.Join(lines, "\n")
}

func (e *SoulEngine) buildPrefetchCache(ctx context.Context, jobID, dialogue, personMD, mapMD string, client llm.Client) {
	maxQ := e.agentCfg.Store.MaxPredictedQuestions
	questions, err := soulagent.RunStorePrefetchQuestionsLLM(ctx, dialogue, personMD, mapMD, maxQ, e.agentCfg, client)
	if err != nil {
		log.Printf("[soul-mcp] prefetch questions LLM: %v", err)
		return
	}
	if len(questions) == 0 {
		return
	}

	all := e.loadAllFacts()
	var blocks []persistence.CacheBlock
	for _, q := range questions {
		selected := recall.Select(all, q, e.agentCfg.Retrieve.MaxFactsInContext, time.Now())
		fb, _ := json.Marshal(selected)
		refs := []string{time.Now().Format("2006-01-02") + ".jsonl"}
		blocks = append(blocks, persistence.CacheBlock{
			Question:   q,
			SourceRefs: refs,
			Content:    string(fb),
		})
	}
	doc := persistence.LLMCacheDoc{
		JobID:              jobID,
		PredictedQuestions: questions,
		Blocks:             blocks,
	}
	doc.AggregateMarkdown = doc.RenderMarkdown()
	_ = e.cache.Write(doc)
}

func (e *SoulEngine) loadAllFacts() []persistence.Fact {
	days := e.agentCfg.Store.MapRecentDays
	if days <= 0 {
		days = 7
	}
	all, _ := e.daily.ListRecentDays(days, time.Now())
	if e.legacy != nil {
		if leg, err := e.legacy.List(); err == nil {
			all = append(all, leg...)
		}
	}
	return all
}
