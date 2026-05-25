# Soul MCP — 架构

> **状态**：**已实现 v4**（`phase=4-async-pipeline`）。本文描述 **本仓库** 内部结构；Host 仅需知工具契约（见 `DESIGN_INTENT.md` §4）。

---

## 1. 系统上下文

```text
┌──────────────── AgentTest (Host) ─────────────────┐
│  WebUI → portal.RunRouterTurn                      │
│    1. soulhook.RetrieveTurn  → soul_retrieve       │
│    2. memoryhook.RetrieveTurn → memory_retrieve    │
│    3. PlanAgent …                                  │
│    N. soulhook.StoreTurn (async) → soul_store      │
└───────────────────────┬───────────────────────────┘
                        │ MCP stdio
                        ▼
┌──────────── AgentTestSoulMCP ──────────────────────┐
│  soul_store  → 异步 4 路 LLM + Go 预取              │
│  soul_retrieve → 快慢双轨 LLM + recall 检索         │
│  只读: soul.agent.yaml                             │
│  可写: storage/history/, person.md, map.md, cache  │
└────────────────────────────────────────────────────┘
```

---

## 2. 进程与部署

| 项 | 说明 |
|----|------|
| 二进制 | `soul-mcp.exe` / `soul-mcp` |
| 协议 | MCP stdio |
| 引擎配置 | `agentConfig/soul-agent.yaml`（`SOUL_MCP_AGENT_CONFIG`） |
| 数据目录 | `SOUL_MCP_DATA_DIR`，默认 `./data` |
| LLM | `SOUL_MCP_LLM_API_BASE` / `SOUL_MCP_API_KEY` / `SOUL_MCP_LLM_MODEL` |
| Agent 灵魂路径 | `SOUL_MCP_SOUL_DOC` → `soul.agent.yaml` |

Host 示例（主项目 `config/app.yaml`，仅引用）：

```yaml
plan_soul_hook:
  enabled: true
  mcp_command: "…/soul-mcp.exe"
  mcp_env:
    SOUL_MCP_DATA_DIR: "…/data"
    SOUL_MCP_LLM_API_BASE: "https://api.deepseek.com"
    SOUL_MCP_LLM_API_KEY: "…"
```

---

## 3. 模块划分

| 包/目录 | 职责 |
|---------|------|
| `cmd/soul-mcp` | MCP 入口、`soul_store` / `soul_retrieve` 注册 |
| `internal/engine` | `SoulEngine`：store/retrieve 编排 |
| `internal/engine/store_pipeline.go` | 异步四路 store |
| `internal/engine/retrieve_pipeline.go` | 快慢双轨 retrieve |
| `internal/soulagent` | 6 段 LLM 提示任务 + hints 格式化 |
| `internal/persistence` | 按天历史、`person.md`、`map.md`、`llm_cache.json` |
| `internal/recall` | 多通道召回 + RRF（时间窗/实体/文本/因果） |
| `internal/config` | `agentConfig` 加载与路径解析 |
| `internal/llm` | OpenAI 兼容 Chat |
| `internal/filter` | 空 context / 闲聊跳过 |
| `internal/response` | JSON 响应封装 |

**不设**：factworld、exec_simple 路由、Memory 工具注册。

---

## 4. Store 管道（异步分治）

```text
soul_store (sync)
    ├─ filter.ShouldSkipStore (可选 skip_chitchat)
    ├─ return { accepted, job_id, phase }
    └─ [async processStore]
            ├─ [并行 wave1]
            │     ├─ Task1 RunStoreDailyLLM    → daily.AppendToday(entries)
            │     ├─ Task2 RunStorePersonLLM   → person.Write
            │     └─ Task3 RunStoreMapLLM      → map.Write
            └─ [wave2]
                  └─ Task4 RunStorePrefetchQuestionsLLM
                        → recall.Select × N 问题
                        → llm_cache.Write
```

无 LLM 时：仅 Task1 降级一条四维骨架事实写入今日文件。

**禁止**：单 LLM 调用同时输出 person + map + entries（违反 S6）。

---

## 5. Retrieve 管道（快慢双轨）

```text
soul_retrieve (sync)
    ├─ 读取 person.md, map.md, llm_cache.json, soul.agent.yaml
    ├─ loadAllFacts() = 最近 N 天按天 JSONL + 可选 legacy history.facts.jsonl
    │
    ├─ [有 LLM]
    │     ├─ Gate LLM (cache + map + person + query + soul)
    │     │     ├─ sufficient → 快通道 hints_markdown
    │     │     └─ else → retrieval_tags { entities, categories, date_hints, keywords }
    │     │              → recall.Select(tags + query)
    │     │              → Compose LLM (facts JSON + map + person + cache + soul)
    │     └─ FormatFinalHints(soul + body)
    │
    └─ [无 LLM] FallbackRetrieveV4 + FormatFinalHints
```

**`retrieval_tags`**：慢通道检索线索，**不是**哲学四维的复刻；Go 用其增强 `recall.Select` 的 query。

---

## 6. 数据布局（运行时）

```text
{SOUL_MCP_DATA_DIR}/
  storage/history/
    2026-05-24.jsonl          # 冷存储：四维事实
    2026-05-25.jsonl
  person.md                   # 用户画像（LLM 任务2）
  map.md                      # 热索引（LLM 任务3）
  llm_cache.json              # 预取缓存（任务4）
  soul.agent.yaml             # 建议放在 exe 同级或 SOUL_MCP_SOUL_DOC
  history.facts.jsonl         # 可选，旧版只读合并检索

agentConfig/
  soul-agent.yaml             # 路径、阈值、6 段 llm.*_system、fact_dimensions
```

### 6.1 按天事实行示例（四维）

```json
{
  "id": "fact-20260524-0",
  "summary": "讨论 Game01 渲染管线锁语义",
  "evidence": "用户：并发锁导致…",
  "stored_at": "2026-05-24T13:00:00Z",
  "phenomenon": {
    "entity": ["Game01", "LibGDX-RenderPipeline"],
    "category": ["Architecture", "CodeBug"],
    "artifacts": ["src/lock/gate.go"]
  },
  "spatiotemporal": {
    "chronos": "2026-05-24T13:00:00Z",
    "kairos": "Critical_Debug",
    "domain": "portal/gateway"
  },
  "causality": {
    "intent": "优化并发锁带来的语义丢失",
    "action": ["read_file"],
    "outcome": "Pitfall_Route",
    "evolution_potential": "Medium"
  },
  "existential": {
    "cognitive_align": "Calibrating",
    "mood_tone": "Focused",
    "persona_shift": ["Macro_Architect"]
  },
  "relations": [{"type": "about", "ref": "Agent-Router-Thesis"}]
}
```

### 6.2 llm_cache.json 示例

```json
{
  "updated_at": "2026-05-25T03:00:00Z",
  "job_id": "soul-job-…",
  "predicted_questions": ["昨天 Game01 锁问题解决了吗", "…"],
  "blocks": [
    {
      "question": "昨天 Game01 锁问题解决了吗",
      "source_refs": ["2026-05-24.jsonl"],
      "content": "[{ …facts json… }]"
    }
  ],
  "aggregate_markdown": "…"
}
```

---

## 7. Host 注入格式（冻结）

Host `plan/soulhook` 从 `soul_retrieve` JSON 取 **`hints`**，与 Memory hints、用户原话拼接（Soul 在前）：

```markdown
{soul_hints 全文}

{memory_hints}

---
用户本轮输入:
{user_input}
```

`hints` 内已含 **Agent 灵魂** 与 **Soul 协作提示** 两节（MCP 内 `FormatFinalHints`）。

---

## 8. 检索子系统（`internal/recall`）

| 通道 | 触发 | 说明 |
|------|------|------|
| 时间 | query 含昨天/上周/ISO 日期 | `chronos` / `stored_at` 日历窗 |
| 实体/范畴 | 词重合 | `phenomenon.entity/category/artifacts` |
| 文本 | 词重合 | 全条 `SearchDocument()` |
| 因果 | 失败/成功等词 | `causality.outcome` 加权 |

多路 **RRF** 融合后取 `retrieve.max_facts_in_context`（默认 24）条给 Compose LLM。

---

## 9. 配置要点（`soul-agent.yaml`）

| 键 | 默认 | 含义 |
|----|------|------|
| `store.max_facts_per_turn` | 12 | 每轮最多写入事实条数 |
| `store.max_predicted_questions` | 5 | 预取问题数上限 |
| `store.map_recent_days` | 7 | map 与预取扫描最近天数 |
| `retrieve.max_hints_runes` | 2000 | 最终 hints 截断 |
| `llm.store_daily_system` 等 | 见 yaml | 六段独立提示词 |

---

## 10. 可观测性

| 信号 | 用途 |
|------|------|
| stderr `[soul-mcp] store daily/person/map/prefetch` | 异步任务成败 |
| `phase` in JSON | 版本：`4-async-pipeline` |
| `hints` 长度 | Host `OnTurnRetrieve hints_len` 日志 |

---

## 11. 安全与隐私

- `data/` 含对话摘录：**不得**提交公开 git。  
- API Key 仅环境变量，不进 yaml。  
- 多租户未实现时勿混用同一 `SOUL_MCP_DATA_DIR`。

---

## 12. 测试

```bash
go test ./...
# Host 边界（主项目）: go run ./scripts/soul_boundary_test
```

---

## 13. 版本

| phase | 含义 |
|-------|------|
| `0-stub` | 占位引擎 |
| `2-llm-triad` | 旧三件套单文件（已 superseded） |
| `4-async-pipeline` | 当前：四路 store + 双轨 retrieve + 按天四维 |
