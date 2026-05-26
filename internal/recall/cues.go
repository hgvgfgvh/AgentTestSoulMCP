package recall

import (
	"regexp"
	"strings"
	"time"
)

// TimeWindow 相对/绝对时间检索窗（本地日历日）。
type TimeWindow struct {
	Start, End time.Time
	Active     bool
}

// QueryCues 从用户 query 解析的检索线索。
type QueryCues struct {
	Time        TimeWindow
	TextQuery   string // 去掉时间词后的文本（供词重合）
	RawQuery    string
	WantPitfall bool
	WantSuccess bool
}

var isoDateRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// ParseCues 解析时间、成败等线索（now 用本地时区做「昨天/今天」）。
func ParseCues(query string, now time.Time) QueryCues {
	q := strings.TrimSpace(query)
	cues := QueryCues{RawQuery: q, TextQuery: stripTimeWords(q)}
	loc := now.Location()
	today := dateOnly(now, loc)

	switch {
	case containsAny(q, "昨天", "昨日"):
		cues.Time = dayWindow(today.AddDate(0, 0, -1), loc)
	case containsAny(q, "前天"):
		cues.Time = dayWindow(today.AddDate(0, 0, -2), loc)
	case containsAny(q, "今天", "今日"):
		cues.Time = dayWindow(today, loc)
	case containsAny(q, "大前天"):
		cues.Time = dayWindow(today.AddDate(0, 0, -3), loc)
	case containsAny(q, "上个月", "上月"):
		y, m, _ := today.Date()
		firstThisMonth := time.Date(y, m, 1, 0, 0, 0, 0, loc)
		cues.Time = TimeWindow{
			Start:  firstThisMonth.AddDate(0, -1, 0),
			End:    firstThisMonth,
			Active: true,
		}
	case containsAny(q, "上周", "上星期"):
		start := today.AddDate(0, 0, -7)
		end := today
		cues.Time = TimeWindow{Start: start, End: end, Active: true}
	case containsAny(q, "这周", "本周", "这星期"):
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := today.AddDate(0, 0, -(weekday - 1))
		cues.Time = TimeWindow{Start: start, End: today.AddDate(0, 0, 1), Active: true}
	case containsAny(q, "最近", "近期"):
		cues.Time = TimeWindow{Start: today.AddDate(0, 0, -7), End: today.AddDate(0, 0, 1), Active: true}
	default:
		if m := isoDateRe.FindString(q); m != "" {
			if d, err := time.ParseInLocation("2006-01-02", m, loc); err == nil {
				cues.Time = dayWindow(d, loc)
			}
		}
	}

	if containsAny(q, "失败", "踩坑", "pitfall", "报错", "出错") {
		cues.WantPitfall = true
	}
	if containsAny(q, "成功", "搞定", "跑通") {
		cues.WantSuccess = true
	}
	return cues
}

func dayWindow(day time.Time, loc *time.Location) TimeWindow {
	d := dateOnly(day, loc)
	return TimeWindow{Start: d, End: d.AddDate(0, 0, 1), Active: true}
}

func dateOnly(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func containsAny(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if sub == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func stripTimeWords(q string) string {
	for _, w := range []string{
		"大前天", "前天", "昨天", "昨日", "今天", "今日",
		"上个月", "上月", "上周", "上星期", "这周", "本周", "这星期",
		"最近", "近期", "什么时候", "哪天",
	} {
		q = strings.ReplaceAll(q, w, " ")
	}
	q = isoDateRe.ReplaceAllString(q, " ")
	return strings.TrimSpace(strings.Join(strings.Fields(q), " "))
}

// InTimeWindow 事实时间是否落在窗口内（半开区间 [Start, End)）。
func InTimeWindow(ft time.Time, w TimeWindow) bool {
	if !w.Active || ft.IsZero() {
		return false
	}
	t := ft
	return !t.Before(w.Start) && t.Before(w.End)
}
