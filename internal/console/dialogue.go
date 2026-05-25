package console

import "strings"

// BuildStoreContent 与 Host soulhook 一致的 WebUI 对话格式。
func BuildStoreContent(turnID, userInput, assistantReply string) string {
	var b strings.Builder
	b.WriteString("[source=soul-mcp-console")
	if t := strings.TrimSpace(turnID); t != "" {
		b.WriteString(" turn=" + t)
	}
	b.WriteString("]\n\n## 用户（WebUI）\n")
	b.WriteString(strings.TrimSpace(userInput))
	b.WriteString("\n\n## 助手（WebUI）\n")
	b.WriteString(strings.TrimSpace(assistantReply))
	return b.String()
}
