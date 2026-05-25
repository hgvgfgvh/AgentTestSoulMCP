package persistence

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFact_SearchDocument_fourDimensions(t *testing.T) {
	f := Fact{
		Summary: "讨论 Game01 渲染管线",
		Phenomenon: PhenomenonTags{
			Entity:   []string{"Game01"},
			Category: []string{"Architecture"},
		},
		Spatiotemporal: SpatiotemporalTags{
			Kairos: "Major_Refactor",
			Domain: "portal/gateway",
		},
		Causality: CausalityTags{
			Outcome: "Success_Route",
		},
		Existential: ExistentialTags{
			CognitiveAlign: "Perfect",
		},
	}
	doc := f.SearchDocument()
	for _, want := range []string{"Game01", "Architecture", "Major_Refactor", "Success_Route", "Perfect"} {
		if !strings.Contains(doc, want) {
			t.Fatalf("SearchDocument missing %q in %q", want, doc)
		}
	}
}

func TestFact_NormalizeLegacy(t *testing.T) {
	f := Fact{
		Tags:     []string{"event"},
		TimeHint: "2026-05-20",
		Entities: map[string][]string{"object": {"Alpha"}},
		Summary:  "legacy row",
	}
	f.NormalizeLegacy()
	if len(f.Phenomenon.Category) == 0 || f.Phenomenon.Entity[0] != "Alpha" {
		t.Fatalf("legacy normalize: %+v", f.Phenomenon)
	}
	if f.Spatiotemporal.Chronos != "2026-05-20" {
		t.Fatalf("chronos=%q", f.Spatiotemporal.Chronos)
	}
}

func TestHistoryStore_Append_setsChronos(t *testing.T) {
	dir := t.TempDir()
	h := NewHistoryStore(filepath.Join(dir, "h.jsonl"))
	_ = h.Append([]Fact{{Summary: "x"}})
	list, _ := h.List()
	if len(list) != 1 || list[0].Spatiotemporal.Chronos == "" {
		t.Fatalf("chronos not set: %+v", list[0])
	}
}
