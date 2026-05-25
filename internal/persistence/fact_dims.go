package persistence

import (
	"encoding/json"
	"strings"
)

// PhenomenonTags 现象与客体维度（实体、范畴、造物留痕）。
type PhenomenonTags struct {
	Entity    []string `json:"entity,omitempty"`
	Category  []string `json:"category,omitempty"`
	Artifacts []string `json:"artifacts,omitempty"`
}

// SpatiotemporalTags 先验时空维度（绝对时序、契机、作用域）。
type SpatiotemporalTags struct {
	Chronos string `json:"chronos,omitempty"`
	Kairos  string `json:"kairos,omitempty"`
	Domain  string `json:"domain,omitempty"`
}

// CausalityTags 行为因果维度（动机、路径、终态、进化潜能）。
type CausalityTags struct {
	Intent             string   `json:"intent,omitempty"`
	Action             []string `json:"action,omitempty"`
	Outcome            string   `json:"outcome,omitempty"`
	EvolutionPotential string   `json:"evolution_potential,omitempty"`
}

// ExistentialTags 存在与互涉维度（认知对齐、情绪、性格投影）。
type ExistentialTags struct {
	CognitiveAlign string   `json:"cognitive_align,omitempty"`
	MoodTone       string   `json:"mood_tone,omitempty"`
	PersonaShift   []string `json:"persona_shift,omitempty"`
}

// SearchDocument 拼接用于 SelectRelevant 词重合的检索文本（含旧版字段兼容）。
func (f Fact) SearchDocument() string {
	var parts []string
	parts = append(parts, f.Summary, f.Evidence, f.TimeHint, f.Source)
	for _, t := range f.Tags {
		parts = append(parts, t)
	}
	for k, vals := range f.Entities {
		parts = append(parts, k)
		parts = append(parts, vals...)
	}
	if p := f.Phenomenon; p.Entity != nil || p.Category != nil || p.Artifacts != nil {
		parts = append(parts, p.Entity...)
		parts = append(parts, p.Category...)
		parts = append(parts, p.Artifacts...)
	}
	if s := f.Spatiotemporal; s.Chronos != "" || s.Kairos != "" || s.Domain != "" {
		parts = append(parts, s.Chronos, s.Kairos, s.Domain)
	}
	if c := f.Causality; c.Intent != "" || c.Outcome != "" || c.EvolutionPotential != "" {
		parts = append(parts, c.Intent, c.Outcome, c.EvolutionPotential)
		parts = append(parts, c.Action...)
	}
	if e := f.Existential; e.CognitiveAlign != "" || e.MoodTone != "" {
		parts = append(parts, e.CognitiveAlign, e.MoodTone)
		parts = append(parts, e.PersonaShift...)
	}
	for _, rel := range f.Relations {
		parts = append(parts, rel.Type, rel.Ref)
	}
	return strings.Join(parts, " ")
}

// NormalizeLegacy 将旧版 tags/entities/time_hint 映射到四维（仅内存，不写回除非 store 重写）。
func (f *Fact) NormalizeLegacy() {
	if f.Phenomenon.Entity == nil && f.Entities != nil {
		if v, ok := f.Entities["person"]; ok {
			f.Phenomenon.Entity = append(f.Phenomenon.Entity, v...)
		}
		if v, ok := f.Entities["object"]; ok {
			f.Phenomenon.Entity = append(f.Phenomenon.Entity, v...)
		}
		if v, ok := f.Entities["event"]; ok {
			f.Phenomenon.Entity = append(f.Phenomenon.Entity, v...)
		}
	}
	if len(f.Phenomenon.Category) == 0 && len(f.Tags) > 0 {
		f.Phenomenon.Category = append([]string(nil), f.Tags...)
	}
	if f.Spatiotemporal.Chronos == "" && f.TimeHint != "" {
		f.Spatiotemporal.Chronos = f.TimeHint
	}
}

// DimensionLabels 用于 hints 展示的紧凑标签行。
func (f Fact) DimensionLabels() string {
	var labels []string
	for _, c := range f.Phenomenon.Category {
		if c != "" {
			labels = append(labels, "Cat:"+c)
		}
	}
	if k := strings.TrimSpace(f.Spatiotemporal.Kairos); k != "" {
		labels = append(labels, "Kairos:"+k)
	}
	if d := strings.TrimSpace(f.Spatiotemporal.Domain); d != "" {
		labels = append(labels, "Domain:"+d)
	}
	if o := strings.TrimSpace(f.Causality.Outcome); o != "" {
		labels = append(labels, "Outcome:"+o)
	}
	if a := strings.TrimSpace(f.Existential.CognitiveAlign); a != "" {
		labels = append(labels, "Align:"+a)
	}
	if len(labels) == 0 && len(f.Tags) > 0 {
		return strings.Join(f.Tags, ",")
	}
	return strings.Join(labels, " · ")
}

// MarshalFactJSON 序列化单条事实（测试/调试）。
func MarshalFactJSON(f Fact) string {
	b, _ := json.Marshal(f)
	return string(b)
}
