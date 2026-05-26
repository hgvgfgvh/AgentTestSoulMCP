package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentConfig agentConfig/soul-agent.yaml — 三件套路径与 LLM 提示。
type AgentConfig struct {
	ID          string `yaml:"id"`
	Version     int    `yaml:"version"`
	Description string `yaml:"description"`

	Files struct {
		HistoryDir string `yaml:"history_dir"`
		History    string `yaml:"history,omitempty"` // 旧版单文件，只读兼容
		Person     string `yaml:"person"`
		Map        string `yaml:"map"`
		LLMCache   string `yaml:"llm_cache"`
		Soul       string `yaml:"soul"`
	} `yaml:"files"`

	// FactDimensions 历史事实四维标签体系（声明式参考，注入 store 提示词）。
	FactDimensions FactDimensionsSchema `yaml:"fact_dimensions"`
	// FactTags 已废弃，仅兼容旧配置；请用 fact_dimensions。
	FactTags []string `yaml:"fact_tags,omitempty"`

	Store struct {
		MaxFactsPerTurn       int  `yaml:"max_facts_per_turn"`
		MaxPredictedQuestions int  `yaml:"max_predicted_questions"`
		MapRecentDays         int  `yaml:"map_recent_days"`
		SkipChitchat          bool `yaml:"skip_chitchat"`
	} `yaml:"store"`

	Retrieve struct {
		MaxFactsInContext int `yaml:"max_facts_in_context"`
		MaxHintsRunes     int `yaml:"max_hints_runes"`
		// DefaultRecentDays 无时间线索时加载的最近日历日数。
		DefaultRecentDays int `yaml:"default_recent_days"`
		// MaxLoadDays 动态按时间窗/date_hints 加载时的日文件数上限。
		MaxLoadDays int `yaml:"max_load_days"`
	} `yaml:"retrieve"`

	LLM struct {
		StoreDailySystem      string `yaml:"store_daily_system"`
		StorePersonSystem     string `yaml:"store_person_system"`
		StoreMapSystem        string `yaml:"store_map_system"`
		StorePrefetchSystem   string `yaml:"store_prefetch_system"`
		RetrieveGateSystem    string `yaml:"retrieve_gate_system"`
		RetrieveComposeSystem string `yaml:"retrieve_compose_system"`
		// 旧版合并提示（兼容）
		StoreSystem    string `yaml:"store_system,omitempty"`
		RetrieveSystem string `yaml:"retrieve_system,omitempty"`
	} `yaml:"llm"`
}

// DimensionField 单维字段说明（agentConfig 声明）。
type DimensionField struct {
	Description string   `yaml:"description"`
	Philosophy  string   `yaml:"philosophy,omitempty"`
	Values      []string `yaml:"values,omitempty"`
	Examples    []string `yaml:"examples,omitempty"`
}

// FactDimensionsSchema 四维标签：现象、时空、因果、存在。
type FactDimensionsSchema struct {
	Phenomenon     map[string]DimensionField `yaml:"phenomenon"`
	Spatiotemporal map[string]DimensionField `yaml:"spatiotemporal"`
	Causality      map[string]DimensionField `yaml:"causality"`
	Existential    map[string]DimensionField `yaml:"existential"`
}

// LoadAgentConfig 读取 agentConfig。
func LoadAgentConfig(path string) (*AgentConfig, error) {
	if strings.TrimSpace(path) == "" {
		path = ResolveAgentConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c AgentConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	c.applyDefaults()
	return &c, nil
}

func (c *AgentConfig) applyDefaults() {
	if c.Files.HistoryDir == "" {
		c.Files.HistoryDir = "storage/history"
	}
	if c.Files.Person == "" {
		c.Files.Person = "person.md"
	}
	if c.Files.Map == "" {
		c.Files.Map = "map.md"
	}
	if c.Files.LLMCache == "" {
		c.Files.LLMCache = "llm_cache.json"
	}
	if c.Files.Soul == "" {
		c.Files.Soul = "soul.agent.yaml"
	}
	if c.Store.MaxFactsPerTurn <= 0 {
		c.Store.MaxFactsPerTurn = 12
	}
	if c.Store.MaxPredictedQuestions <= 0 {
		c.Store.MaxPredictedQuestions = 5
	}
	if c.Store.MapRecentDays <= 0 {
		c.Store.MapRecentDays = 7
	}
	if c.Retrieve.MaxFactsInContext <= 0 {
		c.Retrieve.MaxFactsInContext = 24
	}
	if c.Retrieve.MaxHintsRunes <= 0 {
		c.Retrieve.MaxHintsRunes = 2000
	}
	if c.Retrieve.DefaultRecentDays <= 0 {
		c.Retrieve.DefaultRecentDays = c.Store.MapRecentDays
		if c.Retrieve.DefaultRecentDays <= 0 {
			c.Retrieve.DefaultRecentDays = 7
		}
	}
	if c.Retrieve.MaxLoadDays <= 0 {
		c.Retrieve.MaxLoadDays = 90
	}
}

// ResolveAgentConfigPath agentConfig 路径。
func ResolveAgentConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("SOUL_MCP_AGENT_CONFIG")); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "agentConfig", "soul-agent.yaml")
	}
	return "agentConfig/soul-agent.yaml"
}

// ResolvePaths 兼容旧调用（person + soul；history 为按天目录）。
func (c *AgentConfig) ResolvePaths(dataDir string) (history, person, soul string) {
	p := c.ResolveDataPaths(dataDir)
	return p.HistoryDir, p.Person, p.Soul
}
