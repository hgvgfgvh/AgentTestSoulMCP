package persistence

import (
	"os"
	"path/filepath"
	"strings"
)

// PersonDoc 用户画像 Markdown（LLM 可写）。
type PersonDoc struct {
	path string
}

func NewPersonDoc(path string) *PersonDoc {
	return &PersonDoc{path: path}
}

func (p *PersonDoc) Read() (string, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "# 用户画像\n\n（尚无记录）\n", nil
		}
		return "", err
	}
	return string(data), nil
}

// Write 覆盖 person.yaml，并写 .bak 备份。
func (p *PersonDoc) Write(markdown string) error {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return err
	}
	if old, err := os.ReadFile(p.path); err == nil && len(old) > 0 {
		_ = os.WriteFile(p.path+".bak", old, 0o644)
	}
	return os.WriteFile(p.path, []byte(markdown+"\n"), 0o644)
}
