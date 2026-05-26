package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client OpenAI 兼容 Chat Completions。
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// StoreLLMEnabled store 四路 LLM（默认：已配置 API 则 true；SOUL_MCP_LLM_EXTRACT=0 关闭）。
func StoreLLMEnabled() bool {
	if v := strings.TrimSpace(os.Getenv("SOUL_MCP_LLM_EXTRACT")); v == "0" || strings.EqualFold(v, "false") {
		return false
	}
	_, ok := ConfigFromEnv()
	return ok
}

// RetrieveLLMEnabled retrieve Gate/Compose LLM（默认：已配置 API 则 true；SOUL_MCP_RETRIEVE_LLM=0 关闭）。
func RetrieveLLMEnabled() bool {
	if v := strings.TrimSpace(os.Getenv("SOUL_MCP_RETRIEVE_LLM")); v == "0" || strings.EqualFold(v, "false") {
		return false
	}
	_, ok := ConfigFromEnv()
	return ok
}

// ConfigFromEnv SOUL_MCP_LLM_* 环境变量。
func ConfigFromEnv() (Client, bool) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("SOUL_MCP_LLM_API_BASE")), "/")
	if base == "" {
		return Client{}, false
	}
	key := os.Getenv("SOUL_MCP_LLM_API_KEY")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	model := strings.TrimSpace(os.Getenv("SOUL_MCP_LLM_MODEL"))
	if model == "" {
		model = "deepseek-chat"
	}
	timeout := 45 * time.Second
	if v := os.Getenv("SOUL_MCP_LLM_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}
	return Client{
		BaseURL: base,
		APIKey:  key,
		Model:   model,
		HTTP:    &http.Client{Timeout: timeout},
	}, true
}

// LLMExtractEnabled 同 StoreLLMEnabled（兼容旧名）。
func LLMExtractEnabled() bool { return StoreLLMEnabled() }

// LLMRetrieveComposeEnabled 同 RetrieveLLMEnabled（兼容旧名）。
func LLMRetrieveComposeEnabled() bool { return RetrieveLLMEnabled() }

type chatReq struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

const nowContextSlotMinutes = 30

// floorTo30Min 将时刻归入 30 分钟时间槽起点（利于 LLM 前缀 KV cache 命中）。
func floorTo30Min(t time.Time) time.Time {
	loc := t.Location()
	t = t.In(loc)
	m := t.Minute() - (t.Minute() % nowContextSlotMinutes)
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, 0, 0, loc)
}

// FormatNowContext 每次 LLM 请求的 user 消息前缀：解析「昨天/今天」、chronos 等时以此为准。
// 时间粒度为 30 分钟槽（不含秒），同一槽内多次请求前缀一致。
func FormatNowContext(now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	loc := now.Location()
	local := now.In(loc)
	slotStart := floorTo30Min(local)
	slotEnd := slotStart.Add(nowContextSlotMinutes*time.Minute - time.Minute)
	utcStart := slotStart.UTC()
	utcEnd := slotEnd.UTC()
	day := slotStart.Format("2006-01-02")
	return fmt.Sprintf(
		"## 当前时间节点（系统提供，勿编造；30分钟时间槽）\n"+
			"- 本地时间槽: %s ~ %s\n"+
			"- UTC 时间槽: %s ~ %s\n"+
			"- 时区: %s\n"+
			"- 今日（日历日）: %s\n"+
			"- 昨日: %s\n"+
			"- 前天: %s\n\n",
		slotStart.Format("2006-01-02 15:04"),
		slotEnd.Format("15:04"),
		utcStart.Format("2006-01-02T15:04Z"),
		utcEnd.Format("2006-01-02T15:04Z"),
		loc.String(),
		day,
		slotStart.AddDate(0, 0, -1).Format("2006-01-02"),
		slotStart.AddDate(0, 0, -2).Format("2006-01-02"),
	)
}

// ChatJSON 请求 JSON 回复（user 前自动附带 FormatNowContext）。
func (c *Client) ChatJSON(ctx context.Context, system, user string) (string, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 45 * time.Second}
	}
	user = FormatNowContext(time.Now()) + user
	url := c.BaseURL + "/chat/completions"
	body, _ := json.Marshal(chatReq{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out chatResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("llm: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("llm: empty choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
