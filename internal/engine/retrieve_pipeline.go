package engine

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"AgentTestSoulMCP/internal/llm"
	"AgentTestSoulMCP/internal/recall"
	"AgentTestSoulMCP/internal/soulagent"
)

func (e *SoulEngine) runRetrieve(ctx context.Context, query string) string {
	personMD, _ := e.person.Read()
	mapMD, _ := e.mapDoc.Read()
	soulMD, _ := e.soul.ReadMarkdown()
	cacheDoc, _ := e.cache.Read()
	cacheMD := cacheDoc.RenderMarkdown()

	client, hasLLM := llm.ConfigFromEnv()
	all := e.loadAllFacts()

	if !hasLLM {
		selected := recall.Select(all, query, e.agentCfg.Retrieve.MaxFactsInContext, time.Now())
		return soulagent.FallbackRetrieveV4(query, personMD, mapMD, cacheMD, soulMD, selected, e.agentCfg.Retrieve.MaxHintsRunes)
	}

	gate, err := soulagent.RunRetrieveGateLLM(ctx, query, personMD, mapMD, cacheMD, soulMD, e.agentCfg, client)
	if err != nil {
		log.Printf("[soul-mcp] retrieve gate LLM: %v", err)
	}
	var body string
	if gate.Sufficient && strings.TrimSpace(gate.HintsMarkdown) != "" {
		body = gate.HintsMarkdown
	} else {
		searchQ := query
		if gate.RetrievalTags != nil {
			if extra := gate.RetrievalTags.QueryString(); extra != "" {
				searchQ = query + " " + extra
			}
		}
		selected := recall.Select(all, searchQ, e.agentCfg.Retrieve.MaxFactsInContext, time.Now())
		fb, _ := json.Marshal(selected)
		composed, err := soulagent.RunRetrieveComposeLLM(ctx, query, personMD, mapMD, cacheMD, soulMD, string(fb), e.agentCfg, client)
		if err != nil {
			log.Printf("[soul-mcp] retrieve compose LLM: %v", err)
			body = soulagent.FallbackRetrieveV4(query, personMD, mapMD, cacheMD, soulMD, selected, 0)
		} else if strings.TrimSpace(composed) != "" {
			body = composed
		} else {
			body = soulagent.FallbackRetrieveV4(query, personMD, mapMD, cacheMD, soulMD, selected, 0)
		}
	}
	return soulagent.FormatFinalHints(soulMD, body, e.agentCfg.Retrieve.MaxHintsRunes)
}
