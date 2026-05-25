package persistence

import (
	"strings"
	"time"
)

// FactTime 解析事实的物理时间（chronos 优先，否则 stored_at）。
func FactTime(f Fact) time.Time {
	for _, raw := range []string{f.Spatiotemporal.Chronos, f.StoredAt, f.TimeHint} {
		if t, ok := parseChronos(raw); ok {
			return t
		}
	}
	return time.Time{}
}

func parseChronos(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// FactDateKey 本地日历日期键 YYYY-MM-DD（用于日桶索引）。
func FactDateKey(f Fact, loc *time.Location) string {
	t := FactTime(f)
	if t.IsZero() {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	return t.In(loc).Format("2006-01-02")
}
