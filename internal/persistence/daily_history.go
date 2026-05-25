package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DailyHistoryStore 按天落地 history/YYYY-MM-DD.jsonl。
type DailyHistoryStore struct {
	dir string
}

func NewDailyHistoryStore(dir string) *DailyHistoryStore {
	return &DailyHistoryStore{dir: dir}
}

func (d *DailyHistoryStore) dayPath(day time.Time) string {
	name := day.In(time.Local).Format("2006-01-02") + ".jsonl"
	return filepath.Join(d.dir, name)
}

// AppendDay 向指定日历日文件追加事实。
func (d *DailyHistoryStore) AppendDay(facts []Fact, day time.Time) error {
	if len(facts) == 0 {
		return nil
	}
	path := d.dayPath(day)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	now := day.UTC().Format(time.RFC3339)
	for i := range facts {
		if facts[i].ID == "" {
			facts[i].ID = fmt.Sprintf("fact-%s-%d", day.Format("20060102"), i)
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

// AppendToday 写入今日文件。
func (d *DailyHistoryStore) AppendToday(facts []Fact) error {
	return d.AppendDay(facts, time.Now())
}

func (d *DailyHistoryStore) readFile(path string) ([]Fact, error) {
	f, err := os.Open(path)
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

// ListDay 读取某日全部事实。
func (d *DailyHistoryStore) ListDay(dayKey string) ([]Fact, error) {
	path := filepath.Join(d.dir, dayKey+".jsonl")
	return d.readFile(path)
}

// ListRecentDays 合并最近 N 个日历日文件（含今日）。
func (d *DailyHistoryStore) ListRecentDays(days int, now time.Time) ([]Fact, error) {
	if days <= 0 {
		days = 7
	}
	loc := now.Location()
	var merged []Fact
	for i := 0; i < days; i++ {
		day := dateOnly(now, loc).AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		facts, err := d.ListDay(key)
		if err != nil {
			return nil, err
		}
		merged = append(merged, facts...)
	}
	return merged, nil
}

// RecentDayKeys 返回目录内最近 day 文件名（降序）。
func (d *DailyHistoryStore) RecentDayKeys(max int) ([]string, error) {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var keys []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		keys = append(keys, strings.TrimSuffix(e.Name(), ".jsonl"))
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	if max > 0 && len(keys) > max {
		keys = keys[:max]
	}
	return keys, nil
}

func dateOnly(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}
