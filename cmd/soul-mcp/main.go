// soul-mcp：Soul MCP v4（soul_store / soul_retrieve + 伴生开发对话台）
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/console"
	"AgentTestSoulMCP/internal/engine"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	dataDir      = flag.String("data", "", "SOUL_MCP_DATA_DIR")
	agentCfgPath = flag.String("agent-config", "", "agentConfig/soul-agent.yaml")
	engineKind   = flag.String("engine", "", "soul | stub")
	consoleAddr  = flag.String("console", "", "仅启动开发对话台，如 127.0.0.1:8092")
	httpAddr     = flag.String("http", "", "Streamable HTTP MCP + /console/")
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
		Version: "0.4.0-async-pipeline",
	}, nil)
	registerTools(server, eng)

	consoleSrv, consoleErr := console.NewServer(dir, eng, acPath)
	if consoleErr != nil {
		log.Printf("[soul-mcp] console disabled: %v", consoleErr)
		consoleSrv = nil
	}

	if addr := trim(*consoleAddr); addr != "" && trim(*httpAddr) == "" {
		if consoleSrv == nil {
			log.Fatalf("console: %v", consoleErr)
		}
		log.Printf("[soul-mcp] dev console http://%s/console/", addr)
		if err := http.ListenAndServe(addr, consoleWithRoot(consoleSrv)); err != nil {
			log.Fatalf("console: %v", err)
		}
		return
	}

	if addr := trim(*httpAddr); addr != "" {
		mux := http.NewServeMux()
		mux.Handle("/", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil))
		if consoleSrv != nil {
			consoleSrv.MountPath(mux)
			log.Printf("[soul-mcp] dev console http://%s/console/", addr)
		}
		log.Printf("[soul-mcp] streamable HTTP on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("http: %v", err)
		}
		return
	}

	if consoleSrv != nil {
		startConsoleBackground(consoleSrv, dir)
	}

	t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
	log.Printf("[soul-mcp] stdio transport")
	if err := server.Run(context.Background(), t); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func startConsoleBackground(consoleSrv *console.Server, dataDir string) {
	if envTruthy(os.Getenv("SOUL_MCP_CONSOLE_DISABLE")) {
		log.Printf("[soul-mcp] dev console disabled (SOUL_MCP_CONSOLE_DISABLE)")
		return
	}
	addr := trim(os.Getenv("SOUL_MCP_CONSOLE_LISTEN"))
	if addr == "" {
		addr = "127.0.0.1:8092"
	}
	go func() {
		log.Printf("[soul-mcp] dev console (stdio 伴生) http://%s/console/ data_dir=%s", addr, dataDir)
		if err := http.ListenAndServe(addr, consoleWithRoot(consoleSrv)); err != nil {
			log.Printf("[soul-mcp] dev console exit: %v", err)
		}
	}()
}

func consoleWithRoot(consoleSrv *console.Server) http.Handler {
	mux := http.NewServeMux()
	consoleSrv.MountPath(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/console/", http.StatusFound)
	})
	return mux
}

func envTruthy(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes"
}

func trim(s string) string { return strings.TrimSpace(s) }

func resolveDataDir() string {
	if d := trim(*dataDir); d != "" {
		return d
	}
	if d := trim(os.Getenv("SOUL_MCP_DATA_DIR")); d != "" {
		return d
	}
	return "data"
}

func resolveAgentConfig() string {
	if p := trim(*agentCfgPath); p != "" {
		return p
	}
	return config.ResolveAgentConfigPath()
}

func resolveEngineKind() string {
	if k := trim(*engineKind); k != "" {
		return k
	}
	if k := trim(os.Getenv("SOUL_MCP_ENGINE")); k != "" {
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
		Description: `存入 WebUI 对话。异步：按天事实 + person + map + 预取缓存。`,
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
		Description: `取出协作 hints（预取缓存 + 地图 + 画像 + 快慢双轨 LLM）。`,
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
