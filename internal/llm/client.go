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

// LLMExtractEnabled P1 store LLM。
func LLMExtractEnabled() bool {
	v := strings.TrimSpace(os.Getenv("SOUL_MCP_LLM_EXTRACT"))
	return v == "1" || strings.EqualFold(v, "true")
}

// LLMRetrieveComposeEnabled P1 retrieve 编排。
func LLMRetrieveComposeEnabled() bool {
	v := strings.TrimSpace(os.Getenv("SOUL_MCP_RETRIEVE_LLM"))
	return v == "1" || strings.EqualFold(v, "true")
}

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

// ChatJSON 请求 JSON 回复。
func (c *Client) ChatJSON(ctx context.Context, system, user string) (string, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 45 * time.Second}
	}
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
