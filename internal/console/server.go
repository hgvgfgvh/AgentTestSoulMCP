package console

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"AgentTestSoulMCP/internal/config"
	"AgentTestSoulMCP/internal/engine"
	"AgentTestSoulMCP/internal/persistence"
)

//go:embed web/*
var webFS embed.FS

// Server Soul MCP 开发对话台（进程内直调 Engine，等同 soul_store/soul_retrieve）。
type Server struct {
	dataDir    string
	paths      config.DataPaths
	eng        engine.Engine
	soulEngine *engine.SoulEngine
}

// NewServer 创建控制台。
func NewServer(dataDir string, eng engine.Engine, acPath string) (*Server, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	ac, err := config.LoadAgentConfig(acPath)
	if err != nil {
		return nil, err
	}
	s := &Server{
		dataDir: dataDir,
		paths:   ac.ResolveDataPaths(dataDir),
		eng:     eng,
	}
	if se, ok := eng.(*engine.SoulEngine); ok {
		s.soulEngine = se
	}
	return s, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/retrieve", s.handleRetrieve)
	mux.HandleFunc("/api/store", s.handleStore)
	mux.HandleFunc("/api/turn", s.handleTurn)
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(webRoot)))
	return mux
}

func (s *Server) MountPath(parent *http.ServeMux) {
	parent.Handle("/console/", http.StripPrefix("/console", s.Handler()))
	parent.HandleFunc("/console", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/console/", http.StatusFound)
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	daily := s.paths.HistoryDir
	keys, _ := persistence.NewDailyHistoryStore(daily).RecentDayKeys(7)
	writeJSON(w, http.StatusOK, map[string]any{
		"data_dir":     s.dataDir,
		"history_dir":  daily,
		"history_days": keys,
		"person_bytes": fileSize(s.paths.Person),
		"map_bytes":    fileSize(s.paths.Map),
		"cache_bytes":  fileSize(s.paths.LLMCache),
		"phase":        "4-async-pipeline",
	})
}

type retrieveReq struct {
	Query   string `json:"query"`
	Context string `json:"context"`
}

func (s *Server) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req retrieveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	q := strings.TrimSpace(req.Query)
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query required"})
		return
	}
	ctxStr := strings.TrimSpace(req.Context)
	if ctxStr == "" {
		ctxStr = "用户输入: " + q + "\n"
	}
	raw := s.eng.Retrieve(r.Context(), engine.RetrieveInput{Context: ctxStr, QueryHint: q})
	writeJSON(w, http.StatusOK, parseToolJSON(raw))
}

type storeReq struct {
	UserInput      string `json:"user_input"`
	AssistantReply string `json:"assistant_reply"`
	TurnID         string `json:"turn_id"`
	WaitAsync      bool   `json:"wait_async"`
}

func (s *Server) handleStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req storeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	content := BuildStoreContent(req.TurnID, req.UserInput, req.AssistantReply)
	raw := s.eng.Store(r.Context(), engine.StoreInput{
		Content: content, Source: "soul-mcp-console", Kind: "dialogue",
		CorrelationID: strings.TrimSpace(req.TurnID),
	})
	if req.WaitAsync && s.soulEngine != nil {
		done := make(chan struct{})
		go func() {
			s.soulEngine.WaitAsyncStore()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(90 * time.Second):
		}
	}
	writeJSON(w, http.StatusOK, parseToolJSON(raw))
}

type turnReq struct {
	UserMessage      string `json:"user_message"`
	AssistantMessage string `json:"assistant_message"`
	TurnID           string `json:"turn_id"`
	StoreAfter       bool   `json:"store_after"`
}

// handleTurn 模拟 Host 一轮：先 retrieve（仅用户话），可选再 store（用户+助手）。
func (s *Server) handleTurn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req turnReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	user := strings.TrimSpace(req.UserMessage)
	if user == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_message required"})
		return
	}
	ctx := context.Background()
	retRaw := s.eng.Retrieve(ctx, engine.RetrieveInput{
		Context: "用户输入: " + user + "\n", QueryHint: user,
	})
	out := map[string]any{"retrieve": parseToolJSON(retRaw)}
	assistant := strings.TrimSpace(req.AssistantMessage)
	if req.StoreAfter && assistant != "" {
		storeRaw := s.eng.Store(ctx, engine.StoreInput{
			Content: BuildStoreContent(req.TurnID, user, assistant),
			Source:  "soul-mcp-console", Kind: "dialogue",
			CorrelationID: req.TurnID,
		})
		out["store"] = parseToolJSON(storeRaw)
		if s.soulEngine != nil {
			done := make(chan struct{})
			go func() {
				s.soulEngine.WaitAsyncStore()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(90 * time.Second):
			}
			out["store_completed"] = true
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func parseToolJSON(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	var m map[string]any
	if json.Unmarshal([]byte(raw), &m) == nil {
		return m
	}
	return map[string]any{"raw": raw}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fileSize(path string) int {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(st.Size())
}
