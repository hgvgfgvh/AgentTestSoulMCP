// soul-mcp：Soul MCP（三件套：history.facts.jsonl + person.yaml + soul.agent.yaml）
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/engine"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	dataDir      = flag.String("data", "", "SOUL_MCP_DATA_DIR")
	agentCfgPath = flag.String("agent-config", "", "agentConfig/soul-agent.yaml")
	engineKind   = flag.String("engine", "", "soul | stub")
)

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	dir := resolveDataDir()
	acPath := resolveAgentConfig()
	kind := resolveEngineKind()

	var eng engine.Engine
	var err error
	switch strings.ToLower(kind) {
	case "stub":
		eng, err = engine.NewStubEngine(dir, "")
		log.Printf("[soul-mcp] engine=stub")
	default:
		eng, err = engine.NewSoulEngine(dir, "", acPath)
		if ac, e2 := config.LoadAgentConfig(acPath); e2 == nil && ac != nil {
			paths := ac.ResolveDataPaths(dir)
			log.Printf("[soul-mcp] engine=4-async history=%s person=%s map=%s", paths.HistoryDir, paths.Person, paths.Map)
		}
	}
	if err != nil {
		log.Fatalf("engine: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "agent-test-soul",
		Title:   "AgentTest Soul MCP",
		Version: "0.3.0-llm-triad",
	}, nil)
	registerTools(server, eng)

	t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
	if err := server.Run(context.Background(), t); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func resolveDataDir() string {
	if d := strings.TrimSpace(*dataDir); d != "" {
		return d
	}
	if d := strings.TrimSpace(os.Getenv("SOUL_MCP_DATA_DIR")); d != "" {
		return d
	}
	return "data"
}

func resolveAgentConfig() string {
	if p := strings.TrimSpace(*agentCfgPath); p != "" {
		return p
	}
	return config.ResolveAgentConfigPath()
}

func resolveEngineKind() string {
	if k := strings.TrimSpace(*engineKind); k != "" {
		return k
	}
	if k := strings.TrimSpace(os.Getenv("SOUL_MCP_ENGINE")); k != "" {
		return k
	}
	return "soul"
}

func registerTools(server *mcp.Server, eng engine.Engine) {
	type storeArgs struct {
		Content       string `json:"content" jsonschema:"required,WebUI 对话"`
		Source        string `json:"source,omitempty"`
		Kind          string `json:"kind,omitempty"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "soul_store",
		Description: `存入 WebUI 对话。内部 LLM 拆分为 history.facts.jsonl，并更新 person.yaml。soul.agent.yaml 只读。`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args storeArgs) (*mcp.CallToolResult, any, error) {
		out := eng.Store(ctx, engine.StoreInput{
			Content: args.Content, Source: args.Source, Kind: args.Kind, CorrelationID: args.CorrelationID,
		})
		return textResult(out), nil, nil
	})

	type retrieveArgs struct {
		Context   string `json:"context" jsonschema:"required,用户当前输入上下文"`
		QueryHint string `json:"query_hint,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "soul_retrieve",
		Description: `取出协作 hints。LLM 关联历史事实 + person.yaml + soul.agent.yaml，编排为 Markdown。`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args retrieveArgs) (*mcp.CallToolResult, any, error) {
		out := eng.Retrieve(ctx, engine.RetrieveInput{Context: args.Context, QueryHint: args.QueryHint})
		return textResult(out), nil, nil
	})
}

func textResult(jsonText string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: jsonText}},
	}
}

func init() {
	if _, err := os.Stat("agentConfig/soul-agent.yaml"); err == nil {
		if os.Getenv("SOUL_MCP_AGENT_CONFIG") == "" {
			abs, _ := filepath.Abs("agentConfig/soul-agent.yaml")
			_ = os.Setenv("SOUL_MCP_AGENT_CONFIG", abs)
		}
	}
}
