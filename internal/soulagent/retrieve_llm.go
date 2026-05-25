package soulagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/llm"
	"AgentTestSoulMCP/internal/persistence"
)

type retrieveLLMOutput struct {
	HintsMarkdown string `json:"hints_markdown"`
}

// RunRetrieveLLM 取出：LLM 选材 + 按用户画像编排 hints（含 soul 边界摘录）。
func RunRetrieveLLM(ctx context.Context, query, personMD, soulMD string, facts []persistence.Fact, ac *config.AgentConfig, client llm.Client) (string, error) {
	if ac == nil || strings.TrimSpace(ac.LLM.RetrieveSystem) == "" {
		return "", nil
	}
	fb, _ := json.Marshal(facts)
	user := strings.Join([]string{
		"用户当前输入:\n" + query,
		"\n--- person.yaml ---\n" + personMD,
		"\n--- soul.agent.yaml（只读）---\n" + soulMD,
		"\n--- 候选历史事实 JSON ---\n" + string(fb),
	}, "")
	raw, err := client.ChatJSON(ctx, ac.LLM.RetrieveSystem, user)
	if err != nil {
		return "", err
	}
	raw = extractJSONObject(raw)
	var out retrieveLLMOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.HintsMarkdown), nil
}

// FallbackRetrieve 无 LLM 时模板拼接三件套。
func FallbackRetrieve(query, personMD, soulMD string, facts []persistence.Fact, maxRunes int) string {
	var b strings.Builder
	b.WriteString("## Agent 灵魂（用户定义）\n")
	b.WriteString(soulMD)
	b.WriteString("\n## 用户画像\n")
	b.WriteString(personMD)
	b.WriteString("\n## 相关历史事实\n")
	if len(facts) == 0 {
		b.WriteString("（无匹配记录）\n")
	} else {
		for i, f := range facts {
			b.WriteString(formatFactLine(i+1, f))
		}
	}
	b.WriteString("\n## 当前输入\n")
	b.WriteString(query)
	b.WriteString("\n")
	out := b.String()
	if maxRunes > 0 && len([]rune(out)) > maxRunes {
		r := []rune(out)
		out = string(r[:maxRunes]) + "\n…"
	}
	return out
}

func formatFactLine(n int, f persistence.Fact) string {
	s := strings.TrimSpace(f.Summary)
	if s == "" {
		return ""
	}
	labels := f.DimensionLabels()
	when := strings.TrimSpace(f.Spatiotemporal.Chronos)
	if when == "" {
		when = strings.TrimSpace(f.TimeHint)
	}
	if labels != "" && when != "" {
		return fmt.Sprintf("%d. [%s @%s] %s\n", n, labels, when, s)
	}
	if labels != "" {
		return fmt.Sprintf("%d. [%s] %s\n", n, labels, s)
	}
	if when != "" {
		return fmt.Sprintf("%d. [@%s] %s\n", n, when, s)
	}
	return fmt.Sprintf("%d. %s\n", n, s)
}
