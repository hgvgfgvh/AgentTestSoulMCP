package persistence

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// Fact 历史事实一条（JSONL），按四维标签体系组织。
type Fact struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Evidence string `json:"evidence,omitempty"`
	Source   string `json:"source,omitempty"`
	StoredAt string `json:"stored_at"`

	Phenomenon     PhenomenonTags     `json:"phenomenon,omitempty"`
	Spatiotemporal SpatiotemporalTags `json:"spatiotemporal,omitempty"`
	Causality      CausalityTags      `json:"causality,omitempty"`
	Existential    ExistentialTags    `json:"existential,omitempty"`

	Relations []Relation `json:"relations,omitempty"`

	// 旧版字段（读取兼容；新写入请用四维）
	Tags     []string            `json:"tags,omitempty"`
	TimeHint string              `json:"time_hint,omitempty"`
	Entities map[string][]string `json:"entities,omitempty"`
}

// Relation 事实间关系。
type Relation struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
}

// HistoryStore 单一 history.facts.jsonl。
type HistoryStore struct {
	path string
}

func NewHistoryStore(path string) *HistoryStore {
	return &HistoryStore{path: path}
}

func (h *HistoryStore) Append(facts []Fact) error {
	if len(facts) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepathDir(h.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(h.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range facts {
		if facts[i].ID == "" {
			facts[i].ID = "fact-" + now + "-" + itoa(uint64(i))
		}
		if facts[i].StoredAt == "" {
			facts[i].StoredAt = now
		}
		if strings.TrimSpace(facts[i].Spatiotemporal.Chronos) == "" {
			facts[i].Spatiotemporal.Chronos = facts[i].StoredAt
		}
		facts[i].NormalizeLegacy()
		b, err := json.Marshal(facts[i])
		if err != nil {
			return err
		}
		if _, err := f.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryStore) List() ([]Fact, error) {
	f, err := os.Open(h.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []Fact
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var fact Fact
		if json.Unmarshal([]byte(line), &fact) == nil && strings.TrimSpace(fact.Summary) != "" {
			fact.NormalizeLegacy()
			out = append(out, fact)
		}
	}
	return out, sc.Err()
}

func filepathDir(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		return p[:i]
	}
	return "."
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b [24]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
