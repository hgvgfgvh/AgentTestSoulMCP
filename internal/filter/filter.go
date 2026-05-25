package filter

import "strings"

var chitchat = map[string]struct{}{
	"你好": {}, "您好": {}, "hi": {}, "hello": {}, "谢谢": {}, "在吗": {}, "在么": {},
}

// ShouldSkipStore 是否跳过 store。
func ShouldSkipStore(content string) (bool, string) {
	content = strings.TrimSpace(content)
	if len([]rune(content)) < 16 {
		return true, "content_too_short"
	}
	if idx := strings.Index(content, "## 用户"); idx >= 0 {
		rest := content[idx:]
		if end := strings.Index(rest, "\n\n##"); end > 0 {
			user := strings.TrimSpace(rest[len("## 用户"):end])
			user = strings.TrimPrefix(user, "（WebUI）")
			user = strings.TrimSpace(user)
			if _, ok := chitchat[strings.ToLower(user)]; ok {
				return true, "chitchat"
			}
		}
	}
	return false, ""
}

// ShouldSkipRetrieve 空 context。
func ShouldSkipRetrieve(context string) (bool, string) {
	if len([]rune(strings.TrimSpace(context))) < 2 {
		return true, "empty_context"
	}
	return false, ""
}

// ExtractUserQuery 从 Host context 字符串取用户输入。
func ExtractUserQuery(context string) string {
	context = strings.TrimSpace(context)
	for _, prefix := range []string{"用户输入:", "用户诉求:", "用户本轮输入:"} {
		if i := strings.Index(context, prefix); i >= 0 {
			rest := strings.TrimSpace(context[i+len(prefix):])
			if nl := strings.Index(rest, "\n"); nl >= 0 {
				rest = strings.TrimSpace(rest[:nl])
			}
			if rest != "" {
				return rest
			}
		}
	}
	return context
}
