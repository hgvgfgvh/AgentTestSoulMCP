package soulagent

import (
	"context"
	"encoding/json"
	"strings"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/llm"
	"AgentTestSoulMCP/internal/persistence"
)

type storeLLMOutput struct {
	Facts          []persistence.Fact `json:"facts"`
	PersonMarkdown string             `json:"person_markdown"`
}

// RunStoreLLM 存入：LLM 拆分历史事实 + 更新 person.yaml 正文。
func RunStoreLLM(ctx context.Context, dialogue string, currentPerson string, ac *config.AgentConfig, client llm.Client) (facts []persistence.Fact, personMD string, err error) {
	if ac == nil || strings.TrimSpace(ac.LLM.StoreSystem) == "" {
		return nil, "", nil
	}
	user := "当前 person.yaml:\n" + currentPerson + "\n\n本轮对话:\n" + dialogue
	raw, err := client.ChatJSON(ctx, ac.LLM.StoreSystem, user)
	if err != nil {
		return nil, "", err
	}
	raw = extractJSONObject(raw)
	var out storeLLMOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, "", err
	}
	if ac.Store.MaxFactsPerTurn > 0 && len(out.Facts) > ac.Store.MaxFactsPerTurn {
		out.Facts = out.Facts[:ac.Store.MaxFactsPerTurn]
	}
	return out.Facts, strings.TrimSpace(out.PersonMarkdown), nil
}

// FallbackStore 无 LLM 时追加一条原始摘要事实。
func FallbackStore(dialogue, source string) []persistence.Fact {
	return []persistence.Fact{
		{
			Summary:  preview(dialogue, 400),
			Evidence: preview(dialogue, 300),
			Source:   source,
			Phenomenon: persistence.PhenomenonTags{
				Category: []string{"Topic"},
			},
			Spatiotemporal: persistence.SpatiotemporalTags{
				Domain: "portal/webui",
			},
			Causality: persistence.CausalityTags{
				Outcome:            "Unknown",
				EvolutionPotential: "Low",
			},
			Existential: persistence.ExistentialTags{
				CognitiveAlign: "Calibrating",
			},
		},
	}
}

func preview(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max]) + "…"
}
