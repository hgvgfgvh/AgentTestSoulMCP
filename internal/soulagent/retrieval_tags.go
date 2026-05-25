package soulagent

import "strings"

// RetrievalTags retrieve 慢通道：核心 LLM 输出的检索标签。
type RetrievalTags struct {
	Entities   []string `json:"entities,omitempty"`
	Categories []string `json:"categories,omitempty"`
	DateHints  []string `json:"date_hints,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
}

func (t RetrievalTags) QueryString() string {
	var parts []string
	parts = append(parts, t.Entities...)
	parts = append(parts, t.Categories...)
	parts = append(parts, t.DateHints...)
	parts = append(parts, t.Keywords...)
	return strings.TrimSpace(strings.Join(parts, " "))
}
