package persistence

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SoulDoc 用户定义的 Agent 灵魂（只读）。
type SoulDoc struct {
	path string
}

func NewSoulDoc(path string) *SoulDoc {
	return &SoulDoc{path: path}
}

func (s *SoulDoc) ReadMarkdown() (string, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "（未配置 soul.agent.yaml）\n", nil
		}
		return "", err
	}
	// YAML 包装或纯 Markdown
	var wrap struct {
		Soul string `yaml:"soul"`
	}
	if yaml.Unmarshal(data, &wrap) == nil && strings.TrimSpace(wrap.Soul) != "" {
		return strings.TrimSpace(wrap.Soul) + "\n", nil
	}
	return strings.TrimSpace(string(data)) + "\n", nil
}
