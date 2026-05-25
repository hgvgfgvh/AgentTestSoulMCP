package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DataPaths 运行时落地文件绝对路径。
type DataPaths struct {
	HistoryDir    string
	Person        string
	Map           string
	LLMCache      string
	Soul          string
	LegacyHistory string
}

// ResolveDataPaths 解析 data 目录下各文件路径。
func (c *AgentConfig) ResolveDataPaths(dataDir string) DataPaths {
	dataDir = filepath.Clean(dataDir)
	hDir := strings.TrimSpace(c.Files.HistoryDir)
	if hDir == "" {
		hDir = "storage/history"
	}
	person := strings.TrimSpace(c.Files.Person)
	if person == "" {
		person = "person.md"
	}
	m := strings.TrimSpace(c.Files.Map)
	if m == "" {
		m = "map.md"
	}
	cache := strings.TrimSpace(c.Files.LLMCache)
	if cache == "" {
		cache = "llm_cache.json"
	}
	soul := strings.TrimSpace(c.Files.Soul)
	if soul == "" {
		soul = "soul.agent.yaml"
	}
	soulAbs := strings.TrimSpace(os.Getenv("SOUL_MCP_SOUL_DOC"))
	if soulAbs == "" {
		if exe, err := os.Executable(); err == nil {
			soulAbs = filepath.Join(filepath.Dir(exe), filepath.Clean(soul))
		} else {
			soulAbs = filepath.Clean(soul)
		}
	}
	legacy := ""
	if strings.TrimSpace(c.Files.History) != "" {
		legacy = filepath.Join(dataDir, filepath.Clean(c.Files.History))
	}
	return DataPaths{
		HistoryDir:    filepath.Join(dataDir, filepath.Clean(hDir)),
		Person:        filepath.Join(dataDir, filepath.Clean(person)),
		Map:           filepath.Join(dataDir, filepath.Clean(m)),
		LLMCache:      filepath.Join(dataDir, filepath.Clean(cache)),
		Soul:          soulAbs,
		LegacyHistory: legacy,
	}
}
