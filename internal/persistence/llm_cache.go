package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CacheBlock 单条预取缓存。
type CacheBlock struct {
	Question   string   `json:"question"`
	SourceRefs []string `json:"source_refs,omitempty"`
	Content    string   `json:"content"`
}

// LLMCacheDoc store 阶段写入、retrieve 阶段读取的预取缓存。
type LLMCacheDoc struct {
	UpdatedAt          string       `json:"updated_at"`
	JobID              string       `json:"job_id,omitempty"`
	PredictedQuestions []string     `json:"predicted_questions,omitempty"`
	Blocks             []CacheBlock `json:"blocks,omitempty"`
	AggregateMarkdown  string       `json:"aggregate_markdown,omitempty"`
}

// LLMCacheStore llm_cache.json 读写。
type LLMCacheStore struct {
	path string
}

func NewLLMCacheStore(path string) *LLMCacheStore {
	return &LLMCacheStore{path: path}
}

func (c *LLMCacheStore) Read() (LLMCacheDoc, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return LLMCacheDoc{}, nil
		}
		return LLMCacheDoc{}, err
	}
	var doc LLMCacheDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return LLMCacheDoc{}, err
	}
	return doc, nil
}

func (c *LLMCacheStore) Write(doc LLMCacheDoc) error {
	doc.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, append(b, '\n'), 0o644)
}

// RenderMarkdown 拼接全部预取块供 retrieve 使用。
func (d LLMCacheDoc) RenderMarkdown() string {
	if strings.TrimSpace(d.AggregateMarkdown) != "" {
		return d.AggregateMarkdown
	}
	var b strings.Builder
	for i, blk := range d.Blocks {
		if strings.TrimSpace(blk.Content) == "" {
			continue
		}
		b.WriteString("### 预取 ")
		b.WriteString(fmt.Sprintf("%d", i+1))
		if blk.Question != "" {
			b.WriteString("：")
			b.WriteString(blk.Question)
		}
		b.WriteString("\n")
		b.WriteString(blk.Content)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}
