package persistence

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var mapHistoryRefRe = regexp.MustCompile(`storage/history/(\d{4}-\d{2}-\d{2})\.jsonl`)

// SanitizeMapMarkdown 将 map 中错误的历史指针替换为已存在的最近日文件。
func SanitizeMapMarkdown(mapMD string, historyDir string) string {
	valid := existingHistoryDayKeys(historyDir)
	if len(valid) == 0 {
		return mapMD
	}
	preferred := valid[0]
	return mapHistoryRefRe.ReplaceAllStringFunc(mapMD, func(m string) string {
		sub := mapHistoryRefRe.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		if dayFileExists(historyDir, sub[1]) {
			return m
		}
		return "storage/history/" + preferred + ".jsonl"
	})
}

func existingHistoryDayKeys(historyDir string) []string {
	keys, err := NewDailyHistoryStore(historyDir).RecentDayKeys(14)
	if err != nil || len(keys) == 0 {
		today := time.Now().Format("2006-01-02")
		if dayFileExists(historyDir, today) {
			return []string{today}
		}
		return nil
	}
	return keys
}

func dayFileExists(historyDir, dayKey string) bool {
	p := filepath.Join(historyDir, dayKey+".jsonl")
	st, err := os.Stat(p)
	return err == nil && st.Size() > 0
}

// FormatAvailableHistoryFiles 供 map LLM 提示的可用文件列表。
func FormatAvailableHistoryFiles(historyDir string) string {
	keys := existingHistoryDayKeys(historyDir)
	if len(keys) == 0 {
		return "（今日尚无落地文件，指针请写 storage/history/" + time.Now().Format("2006-01-02") + ".jsonl）"
	}
	var lines []string
	for _, k := range keys {
		lines = append(lines, "storage/history/"+k+".jsonl")
	}
	return strings.Join(lines, "\n")
}
