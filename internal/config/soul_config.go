package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SoulBase soul.config 基座（只读）。
type SoulBase struct {
	Version     int    `yaml:"version"`
	ID          string `yaml:"id"`
	DisplayName string `yaml:"display_name"`
	Persona     struct {
		Role       string   `yaml:"role"`
		Tone       []string `yaml:"tone"`
		Boundaries []string `yaml:"boundaries"`
	} `yaml:"persona"`
	Defaults struct {
		Language    string `yaml:"language"`
		ReplyLength string `yaml:"reply_length"`
		Formality   string `yaml:"formality"`
	} `yaml:"defaults"`
	PromptBlocks struct {
		PersonaHeader string `yaml:"persona_header"`
		EventsHeader  string `yaml:"events_header"`
	} `yaml:"prompt_blocks"`
}

// SoulOverlay data/soul_overlay/active.yml 增量人格（可审计）。
type SoulOverlay struct {
	Reason    string `yaml:"reason"`
	CreatedAt string `yaml:"created_at"`
	Patch     struct {
		ToneNotes string `yaml:"tone_notes"`
		ExtraRole string `yaml:"extra_role"`
	} `yaml:"patch"`
}

// LoadSoulBase 加载 soul.config。
func LoadSoulBase(path string) (*SoulBase, error) {
	if strings.TrimSpace(path) == "" {
		path = ResolveSoulConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b SoulBase
	if err := yaml.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	if b.PromptBlocks.PersonaHeader == "" {
		b.PromptBlocks.PersonaHeader = "## 协作人格（Soul）"
	}
	if b.PromptBlocks.EventsHeader == "" {
		b.PromptBlocks.EventsHeader = "## 近期议题与事件（Soul）"
	}
	return &b, nil
}

// LoadOverlay 若不存在返回 nil, nil。
func LoadOverlay(dataDir string) (*SoulOverlay, error) {
	p := filepath.Join(dataDir, "soul_overlay", "active.yml")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var o SoulOverlay
	if err := yaml.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// ResolveSoulConfigPath 解析 soul.config 路径。
func ResolveSoulConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("SOUL_MCP_CONFIG")); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "soul.config")
	}
	return "soul.config"
}
