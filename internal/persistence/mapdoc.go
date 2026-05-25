package persistence

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultMapTemplate = `# Soul 地图（热索引）

> 高价值摘要 + 落地文件指针。由 store 异步任务维护。

## 条目

（尚无记录）
`

// MapDoc 地图文档（热索引层）。
type MapDoc struct {
	path string
}

func NewMapDoc(path string) *MapDoc {
	return &MapDoc{path: path}
}

func (m *MapDoc) Read() (string, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultMapTemplate, nil
		}
		return "", err
	}
	return string(data), nil
}

func (m *MapDoc) Write(markdown string) error {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	if old, err := os.ReadFile(m.path); err == nil && len(old) > 0 {
		_ = os.WriteFile(m.path+".bak", old, 0o644)
	}
	return os.WriteFile(m.path, []byte(markdown+"\n"), 0o644)
}
