package recall

import (
	"strings"
	"time"

	"AgentTestSoulMCP/internal/persistence"
	"AgentTestSoulMCP/internal/textutil"
)

// Select 多通道召回 + RRF 融合（时间 / 实体范畴 / 文本 / 因果终态）。
func Select(facts []persistence.Fact, query string, topN int, now time.Time) []persistence.Fact {
	if len(facts) == 0 {
		return nil
	}
	if topN <= 0 {
		topN = 24
	}
	cues := ParseCues(query, now)
	loc := now.Location()

	textQ := cues.TextQuery
	if textQ == "" {
		textQ = cues.RawQuery
	}

	var rankings [][]int

	// 通道 1：时间窗（有线索时，窗内事实按时间倒序，不要求词重合）
	if cues.Time.Active {
		rankings = append(rankings, rankByTimeWindow(facts, cues.Time, loc))
	}

	// 通道 2：实体 / 范畴 / 造物
	if r := rankByEntity(textQ, facts); len(r) > 0 {
		rankings = append(rankings, r)
	}

	// 通道 3：全文词重合
	if r := rankByText(textQ, facts); len(r) > 0 {
		rankings = append(rankings, r)
	}
	if textQ != cues.RawQuery {
		if r := rankByText(cues.RawQuery, facts); len(r) > 0 {
			rankings = append(rankings, r)
		}
	}

	// 通道 4：因果终态（失败/成功问法）
	if cues.WantPitfall || cues.WantSuccess {
		if r := rankByOutcome(facts, cues.WantPitfall, cues.WantSuccess); len(r) > 0 {
			rankings = append(rankings, r)
		}
	}

	// 无任一路命中且有时间窗：仍返回时间窗内全部（支持纯「昨天问了啥」）
	if len(rankings) == 0 && cues.Time.Active {
		rankings = append(rankings, rankByTimeWindow(facts, cues.Time, loc))
	}

	// 仍无：退回纯文本；再无则最近 N 条
	if len(rankings) == 0 {
		if r := rankByText(cues.RawQuery, facts); len(r) > 0 {
			rankings = append(rankings, r)
		} else {
			rankings = append(rankings, rankByRecency(facts, loc))
		}
	}

	indices := fuseRRF(rankings, topN, 60)
	out := make([]persistence.Fact, 0, len(indices))
	for _, i := range indices {
		if i >= 0 && i < len(facts) {
			out = append(out, facts[i])
		}
	}
	return out
}

func rankByTimeWindow(facts []persistence.Fact, w TimeWindow, loc *time.Location) []int {
	var in []timePair
	for i, f := range facts {
		ft := persistence.FactTime(f)
		if InTimeWindow(ft.In(loc), w) {
			in = append(in, timePair{i, ft})
		}
	}
	sortPairsByTimeDesc(in)
	out := make([]int, len(in))
	for i, p := range in {
		out[i] = p.idx
	}
	return out
}

func rankByEntity(query string, facts []persistence.Fact) []int {
	if strings.TrimSpace(query) == "" {
		return nil
	}
	var ranked []scorePair
	for i, f := range facts {
		doc := strings.Join(f.Phenomenon.Entity, " ") + " " +
			strings.Join(f.Phenomenon.Category, " ") + " " +
			strings.Join(f.Phenomenon.Artifacts, " ")
		sc := textutil.OverlapScore(query, doc)
		if sc > 0 {
			ranked = append(ranked, scorePair{i, sc})
		}
	}
	return sortByScore(ranked)
}

func rankByText(query string, facts []persistence.Fact) []int {
	if strings.TrimSpace(query) == "" {
		return nil
	}
	var ranked []scorePair
	for i, f := range facts {
		sc := textutil.OverlapScore(query, f.SearchDocument())
		if sc > 0 {
			ranked = append(ranked, scorePair{i, sc})
		}
	}
	return sortByScore(ranked)
}

func rankByOutcome(facts []persistence.Fact, pitfall, success bool) []int {
	var ranked []scorePair
	for i, f := range facts {
		o := strings.ToLower(f.Causality.Outcome)
		sc := 0.0
		if pitfall && (strings.Contains(o, "pitfall") || strings.Contains(o, "fail") || strings.Contains(o, "aborted")) {
			sc = 1
		}
		if success && strings.Contains(o, "success") {
			sc = 1
		}
		if sc > 0 {
			ranked = append(ranked, scorePair{i, sc})
		}
	}
	return sortByScore(ranked)
}

func rankByRecency(facts []persistence.Fact, loc *time.Location) []int {
	var all []timePair
	for i, f := range facts {
		all = append(all, timePair{i, persistence.FactTime(f).In(loc)})
	}
	sortPairsByTimeDesc(all)
	out := make([]int, len(all))
	for i, p := range all {
		out[i] = p.idx
	}
	return out
}

type scorePair struct {
	idx int
	sc  float64
}

func sortByScore(ranked []scorePair) []int {
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].sc > ranked[i].sc {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	out := make([]int, len(ranked))
	for i, s := range ranked {
		out[i] = s.idx
	}
	return out
}

type timePair struct {
	idx int
	t   time.Time
}

func sortPairsByTimeDesc(pairs []timePair) {
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].t.After(pairs[i].t) {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
}
