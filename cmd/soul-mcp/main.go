// soul-mcp：人格/议题 Soul MCP Server（Phase-0 骨架）。
//
// 对外工具：soul_store、soul_retrieve（字符串协议）。
// 内部：StubEngine；profile/events 整理 Agent 未实现。
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"AgentTestSoulMCP/internal/engine"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	dataDir    = flag.String("data", "", "数据目录（默认 ./data 或 SOUL_MCP_DATA_DIR）")
	configPath = flag.String("config", "", "soul.config 路径（默认同目录 soul.config）")
)

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	dir := strings.TrimSpace(*dataDir)
	if dir == "" {
		dir = strings.TrimSpace(os.Getenv("SOUL_MCP_DATA_DIR"))
	}
	if dir == "" {
		dir = "data"
	}
	cfg := strings.TrimSpace(*configPath)
	if cfg == "" {
		cfg = strings.TrimSpace(os.Getenv("SOUL_MCP_CONFIG"))
	}
	if cfg == "" {
		cfg = "soul.config"
	}
	if !filepath.IsAbs(cfg) {
		if exe, err := os.Executable(); err == nil {
			cfg = filepath.Join(filepath.Dir(exe), cfg)
		}
	}

	eng, err := engine.NewStubEngine(dir, cfg)
	if err != nil {
		log.Fatalf("engine: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-test-soul",
		Title:   "AgentTest Soul MCP",
		Version: "0.1.0-phase0",
	}, nil)

	registerTools(server, eng)

	t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
	log.Printf("[soul-mcp] stdio transport data_dir=%s config=%s phase=0-stub", dir, cfg)
	if err := server.Run(context.Background(), t); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func registerTools(server *mcp.Server, eng engine.Engine) {
	type storeArgs struct {
		Content       string `json:"content" jsonschema:"required,WebUI 对话等材料（字符串）"`
		Source        string `json:"source,omitempty" jsonschema:"Host 标识，如 agenttest-webui"`
		Kind          string `json:"kind,omitempty" jsonschema:"粗分类，如 dialogue"`
		CorrelationID string `json:"correlation_id,omitempty" jsonschema:"Host 关联 ID，如 turn_id"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name: "soul_store",
		Description: `存入 WebUI 对话等人格/议题材料（字符串协议）。同步 ACK，内部异步整理（Phase-0 仅追加 stub 日志）。
返回 JSON 字符串：accepted、job_id、skipped 等。`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args storeArgs) (*mcp.CallToolResult, any, error) {
		out := eng.Store(ctx, engine.StoreInput{
			Content:       args.Content,
			Source:        args.Source,
			Kind:          args.Kind,
			CorrelationID: args.CorrelationID,
		})
		return textResult(out), nil, nil
	})

	type retrieveArgs struct {
		Context   string `json:"context" jsonschema:"required,Host 本轮上下文字符串（含用户输入）"`
		QueryHint string `json:"query_hint,omitempty" jsonschema:"可选补充检索意图"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name: "soul_retrieve",
		Description: `取出人格/议题参考 hints（字符串协议）。须传 context；同步、快速。Phase-0 为 stub 占位。
返回 JSON 字符串：hints、skipped 等。不得包含 exec_simple_match 等 Memory 路由字段。`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args retrieveArgs) (*mcp.CallToolResult, any, error) {
		out := eng.Retrieve(ctx, engine.RetrieveInput{
			Context:   args.Context,
			QueryHint: args.QueryHint,
		})
		return textResult(out), nil, nil
	})
}

func textResult(jsonText string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: jsonText},
		},
	}
}
