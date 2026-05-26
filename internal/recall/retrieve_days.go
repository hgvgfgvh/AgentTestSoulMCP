package recall

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"AgentTestSoulMCP/internal/soulagent"
)

var (
	isoDayRe   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	isoMonthRe = regexp.MustCompile(`^\d{4}-\d{2}$`)
)

// ResolveRetrieveDayKeys 按 query 时间线索 + Gate date_hints 决定要读取的落地日文件键。
// 无时间线索时退回最近 defaultDays 个日历日；最多加载 maxDays 个文件（超出则保留窗口内最新的 maxDays 天）。
func ResolveRetrieveDayKeys(query string, tags *soulagent.RetrievalTags, now time.Time, defaultDays, maxDays int) []string {
	if defaultDays <= 0 {
		defaultDays = 7
	}
	if maxDays <= 0 {
		maxDays = 90
	}
	loc := now.Location()
	keySet := map[string]struct{}{}

	cues := ParseCues(query, now)
	for _, k := range dayKeysFromTimeWindow(cues.Time, loc) {
		keySet[k] = struct{}{}
	}
	if tags != nil {
		for _, hint := range tags.DateHints {
			for _, k := range dayKeysFromDateHint(strings.TrimSpace(hint), loc) {
				keySet[k] = struct{}{}
			}
		}
	}

	if len(keySet) == 0 {
		return recentCalendarDayKeys(now, defaultDays, loc)
	}
	keys := sortedDayKeys(keySet)
	if len(keys) > maxDays {
		keys = keys[len(keys)-maxDays:]
	}
	return keys
}

func dayKeysFromTimeWindow(w TimeWindow, loc *time.Location) []string {
	if !w.Active {
		return nil
	}
	var keys []string
	for d := dateOnly(w.Start, loc); d.Before(w.End); d = d.AddDate(0, 0, 1) {
		keys = append(keys, d.Format("2006-01-02"))
	}
	return keys
}

func dayKeysFromDateHint(hint string, loc *time.Location) []string {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return nil
	}
	if isoDayRe.MatchString(hint) {
		if t, err := time.ParseInLocation("2006-01-02", hint, loc); err == nil {
			return []string{t.Format("2006-01-02")}
		}
		return nil
	}
	if isoMonthRe.MatchString(hint) {
		t, err := time.ParseInLocation("2006-01-02", hint+"-01", loc)
		if err != nil {
			return nil
		}
		end := t.AddDate(0, 1, 0)
		return dayKeysFromTimeWindow(TimeWindow{Start: t, End: end, Active: true}, loc)
	}
	// 兼容 Gate 在 date_hints 里写「上个月」等
	cues := ParseCues(hint, time.Now().In(loc))
	return dayKeysFromTimeWindow(cues.Time, loc)
}

func recentCalendarDayKeys(now time.Time, days int, loc *time.Location) []string {
	today := dateOnly(now, loc)
	var keys []string
	for i := 0; i < days; i++ {
		keys = append(keys, today.AddDate(0, 0, -i).Format("2006-01-02"))
	}
	return keys
}

func sortedDayKeys(keySet map[string]struct{}) []string {
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
