package soulagent

import (
	"context"
	"encoding/json"
	"strings"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/llm"
	"AgentTestSoulMCP/internal/persistence"
)

type retrieveGateOutput struct {
	Sufficient    bool           `json:"sufficient"`
	HintsMarkdown string         `json:"hints_markdown,omitempty"`
	RetrievalTags *RetrievalTags `json:"retrieval_tags,omitempty"`
}

// RunRetrieveGateLLM 取出第一遍：快通道判定或输出检索标签。
func RunRetrieveGateLLM(ctx context.Context, query, personMD, mapMD, cacheMD, soulMD string, ac *config.AgentConfig, client llm.Client) (retrieveGateOutput, error) {
	sys := ac.LLM.RetrieveGateSystem
	if strings.TrimSpace(sys) == "" {
		return retrieveGateOutput{}, nil
	}
	user := strings.Join([]string{
		"用户当前输入:\n" + query,
		"\n--- LLM 预取缓存 ---\n" + cacheMD,
		"\n--- 地图文档 map.md ---\n" + mapMD,
		"\n--- 用户画像 person ---\n" + personMD,
		"\n--- Agent 灵魂 soul（只读，最终输出须保持其基调）---\n" + soulMD,
	}, "")
	raw, err := client.ChatJSON(ctx, sys, user)
	if err != nil {
		return retrieveGateOutput{}, err
	}
	raw = extractJSONObject(raw)
	var out retrieveGateOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return retrieveGateOutput{}, err
	}
	return out, nil
}

type retrieveComposeOutput struct {
	HintsMarkdown string `json:"hints_markdown"`
}

// RunRetrieveComposeLLM 取出第二遍：慢通道总结。
func RunRetrieveComposeLLM(ctx context.Context, query, personMD, mapMD, cacheMD, soulMD, retrievedJSON string, ac *config.AgentConfig, client llm.Client) (string, error) {
	sys := ac.LLM.RetrieveComposeSystem
	if strings.TrimSpace(sys) == "" {
		sys = ac.LLM.RetrieveSystem
	}
	if strings.TrimSpace(sys) == "" {
		return "", nil
	}
	user := strings.Join([]string{
		"用户当前输入:\n" + query,
		"\n--- 检索到的落地事实 JSON ---\n" + retrievedJSON,
		"\n--- 地图文档 ---\n" + mapMD,
		"\n--- 用户画像 ---\n" + personMD,
		"\n--- LLM 预取缓存 ---\n" + cacheMD,
		"\n--- Agent 灵魂（只读）---\n" + soulMD,
	}, "")
	raw, err := client.ChatJSON(ctx, sys, user)
	if err != nil {
		return "", err
	}
	raw = extractJSONObject(raw)
	var out retrieveComposeOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		// 兼容仅 hints_markdown 或纯文本
		if strings.Contains(raw, "hints_markdown") {
			return "", err
		}
		return strings.TrimSpace(raw), nil
	}
	return strings.TrimSpace(out.HintsMarkdown), nil
}

// FormatFinalHints 最终 hints：灵魂 + Soul 总结正文。
func FormatFinalHints(soulMD, body string, maxRunes int) string {
	var b strings.Builder
	b.WriteString("## Agent 灵魂（用户定义·只读）\n")
	b.WriteString(strings.TrimSpace(soulMD))
	b.WriteString("\n\n## Soul 协作提示\n")
	b.WriteString(strings.TrimSpace(body))
	b.WriteString("\n")
	out := b.String()
	if maxRunes > 0 && len([]rune(out)) > maxRunes {
		r := []rune(out)
		out = string(r[:maxRunes]) + "\n…"
	}
	return out
}

// FallbackRetrieveV4 无 LLM 时模板。
func FallbackRetrieveV4(query, personMD, mapMD, cacheMD, soulMD string, facts []persistence.Fact, maxRunes int) string {
	var b strings.Builder
	if strings.TrimSpace(mapMD) != "" {
		b.WriteString("## 地图\n")
		b.WriteString(mapMD)
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(cacheMD) != "" {
		b.WriteString("## 预取缓存\n")
		b.WriteString(cacheMD)
		b.WriteString("\n\n")
	}
	b.WriteString("## 用户画像\n")
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
	return FormatFinalHints(soulMD, b.String(), maxRunes)
}
