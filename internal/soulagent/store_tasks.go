package soulagent

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/llm"
	"AgentTestSoulMCP/internal/persistence"
)

type storeDailyOutput struct {
	Entries []persistence.Fact `json:"entries"`
}

// RunStoreDailyLLM 任务1：按天落地文档 + 标签。
func RunStoreDailyLLM(ctx context.Context, dialogue, dayKey string, ac *config.AgentConfig, client llm.Client) ([]persistence.Fact, error) {
	sys := ac.LLM.StoreDailySystem
	if strings.TrimSpace(sys) == "" {
		sys = ac.LLM.StoreSystem
	}
	if strings.TrimSpace(sys) == "" {
		return nil, nil
	}
	user := "落地日文件: " + dayKey + ".jsonl\n\n本轮对话:\n" + dialogue
	raw, err := client.ChatJSON(ctx, sys, user)
	if err != nil {
		return nil, err
	}
	raw = extractJSONObject(raw)
	var out storeDailyOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		// 兼容旧版 facts 字段
		var leg struct {
			Facts []persistence.Fact `json:"facts"`
		}
		if json.Unmarshal([]byte(raw), &leg) == nil {
			out.Entries = leg.Facts
		} else {
			return nil, err
		}
	}
	if ac.Store.MaxFactsPerTurn > 0 && len(out.Entries) > ac.Store.MaxFactsPerTurn {
		out.Entries = out.Entries[:ac.Store.MaxFactsPerTurn]
	}
	return out.Entries, nil
}

type storePersonOutput struct {
	PersonMarkdown string `json:"person_markdown"`
}

// RunStorePersonLLM 任务2：更新用户画像 Markdown。
func RunStorePersonLLM(ctx context.Context, dialogue, currentPerson string, ac *config.AgentConfig, client llm.Client) (string, error) {
	sys := ac.LLM.StorePersonSystem
	if strings.TrimSpace(sys) == "" {
		return "", nil
	}
	user := "当前用户画像:\n" + currentPerson + "\n\n本轮对话:\n" + dialogue
	raw, err := client.ChatJSON(ctx, sys, user)
	if err != nil {
		return "", err
	}
	raw = extractJSONObject(raw)
	var out storePersonOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.PersonMarkdown), nil
}

type storeMapOutput struct {
	MapMarkdown string `json:"map_markdown"`
}

// RunStoreMapLLM 任务3：维护地图热索引。
func RunStoreMapLLM(ctx context.Context, dialogue, currentMap, recentDaysSummary string, ac *config.AgentConfig, client llm.Client) (string, error) {
	sys := ac.LLM.StoreMapSystem
	if strings.TrimSpace(sys) == "" {
		return "", nil
	}
	user := strings.Join([]string{
		"当前地图文档:\n" + currentMap,
		"\n最近落地日文件摘要:\n" + recentDaysSummary,
		"\n本轮对话:\n" + dialogue,
	}, "")
	raw, err := client.ChatJSON(ctx, sys, user)
	if err != nil {
		return "", err
	}
	raw = extractJSONObject(raw)
	var out storeMapOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.MapMarkdown), nil
}

type storePrefetchQuestionsOutput struct {
	Questions []string `json:"questions"`
}

// RunStorePrefetchQuestionsLLM 任务4a：预测后续问题。
func RunStorePrefetchQuestionsLLM(ctx context.Context, dialogue, personMD, mapMD string, maxQ int, ac *config.AgentConfig, client llm.Client) ([]string, error) {
	sys := ac.LLM.StorePrefetchSystem
	if strings.TrimSpace(sys) == "" {
		return nil, nil
	}
	user := strings.Join([]string{
		"用户画像:\n" + personMD,
		"\n地图文档:\n" + mapMD,
		"\n最近对话:\n" + dialogue,
		"\n请预测用户最可能继续问的 " + strconv.Itoa(maxQ) + " 个问题。",
	}, "")
	raw, err := client.ChatJSON(ctx, sys, user)
	if err != nil {
		return nil, err
	}
	raw = extractJSONObject(raw)
	var out storePrefetchQuestionsOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if maxQ > 0 && len(out.Questions) > maxQ {
		out.Questions = out.Questions[:maxQ]
	}
	return out.Questions, nil
}
