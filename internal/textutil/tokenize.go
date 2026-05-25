package textutil

import (
	"strings"
	"unicode"
)

// Tokens 中英数 token 集合。
func Tokens(s string) map[string]int {
	out := map[string]int{}
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		t := strings.ToLower(b.String())
		if len(t) >= 1 {
			out[t]++
		}
		b.Reset()
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

// OverlapScore 查询与文档词重合分 [0,1]。
func OverlapScore(query, doc string) float64 {
	q := Tokens(query)
	if len(q) == 0 {
		return 0
	}
	d := Tokens(doc)
	if len(d) == 0 {
		return 0
	}
	hit := 0
	for t := range q {
		if d[t] > 0 {
			hit++
		}
	}
	return float64(hit) / float64(len(q))
}
