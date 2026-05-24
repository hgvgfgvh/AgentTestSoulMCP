# Soul MCP — 架构

> 状态：**规划**（2026-05-24）。下文为目标架构；实现后须同步 `IMPLEMENTATION_PROGRESS.md` 与 `ARCHITECTURE_DRIFT.md`。

---

## 1. 系统上下文

```text
┌──────────────── AgentTest (Host) ─────────────────┐
│  WebUI → portal.RunRouterTurn                      │
│    1. soulhook.RetrieveTurn  → stdio soul_retrieve │
│    2. memoryhook.RetrieveTurn → memory_retrieve    │
│    3. PlanAgent / Affective                        │
│    …                                               │
│    N. soulhook.StoreTurn (async) → soul_store      │
│    N. memoryhook.StoreTurn (async) → memory_store  │
└───────────────────────┬───────────────────────────┘
                        │ stdio MCP
                        ▼
┌──────────── AgentTestSoulMCP (本进程) ─────────────┐
│  MCP Server: soul_store | soul_retrieve            │
│  soul.config (只读基座)                             │
│  data/: profile.jsonl, events.jsonl, episodes/,    │
│         soul_overlay/, queue/                      │
│  internal/: pipeline, template, optional LLM     │
└────────────────────────────────────────────────────┘
```

---

## 2. 进程与部署

| 项 | 规划 |
|----|------|
| 二进制名 | `soul-mcp.exe`（Windows）/ `soul-mcp`（Unix） |
| 协议 | MCP stdio，与 `memory-mcp` 相同模式 |
| 配置 | 环境变量 `SOUL_*`；旁路文件 `soul.config` |
| 数据目录 | `SOUL_DATA_DIR` 或默认 `./data` |

Host `config/app.yaml` 规划段：

```yaml
plan_soul_hook:
  enabled: true
  mcp_command: ["C:/DATA/GODATA/AgentTestSoulMCP/soul-mcp.exe"]
  mcp_env:
    SOUL_DATA_DIR: "..."
    # SOUL_STORE_LLM_*  optional
    # SOUL_RETRIEVE_LLM: "0"
```

---

## 3. 模块划分（实现时）

| 包/目录 | 职责 |
|---------|------|
| `cmd/soul-mcp` | MCP 入口、工具注册 |
| `internal/mcp` | `soul_store` / `soul_retrieve` handler |
| `internal/pipeline` | store 队列、job 消费 |
| `internal/profile` | profile.jsonl CRUD、去重、confidence 衰减 |
| `internal/events` | events.jsonl、议题合并、指代索引 |
| `internal/template` | retrieve 无 LLM 组装 |
| `internal/llm` | 可选 store 整理 / retrieve compose |
| `internal/config` | 加载 `soul.config` + overlay |

**刻意不设**：`internal/graph` 执行经验图、factworld、BM25 路由（属 Memory）。

---

## 4. Store 管道

```text
soul_store (sync)
    ├─ validate JSON → enqueue job → return accepted
    └─ [async worker]
            ├─ append raw → data/episodes/{episode_id}.json
            ├─ optional LLM extract (profile deltas + events)
            ├─ rules fallback if no LLM
            ├─ merge profile (dedupe by key)
            ├─ upsert events (same topic → update last_mentioned)
            └─ optional: write overlay *suggestion* (not auto-apply)
```

**LLM 在 store**：仅 MCP 内部；Host **不等待**整理完成（S2）。

---

## 5. Retrieve 管道

```text
soul_retrieve (sync, budgeted)
    ├─ load soul.config (readonly)
    ├─ load soul_overlay/active.yml if exists
    ├─ match profile (keyword / optional embedding later)
    ├─ match events (user input + context BM25 or tag overlap)
    ├─ [optional] LLM compose if SOUL_RETRIEVE_LLM=1 && within timeout
    └─ template → persona_prompt + event_context + retrieve_meta
```

**默认路径**：模板组装（S6），P99 目标 &lt; 50ms（无 LLM、本地 data）。

---

## 6. 数据布局

```text
AgentTestSoulMCP/
  soul.config              # 基座（git 可跟踪 example，生产本地）
  soul.config.example
  data/                    # 运行时（gitignore）
    profile.jsonl
    events.jsonl
    episodes/
      {episode_id}.json
    soul_overlay/
      active.yml
      suggestions/
    queue/
      pending.jsonl
```

### 6.1 profile.jsonl 行示例

```json
{
  "key": "user.preferred_name",
  "value": "老王",
  "confidence": 0.9,
  "source_episode_id": "ep-uuid",
  "updated_at": "2026-05-22T10:00:00Z"
}
```

### 6.2 events.jsonl 行示例

```json
{
  "event_id": "evt-uuid",
  "kind": "paper",
  "title": "某某架构论文",
  "summary": "讨论结论与实验设置",
  "entities": ["项目 Alpha"],
  "last_mentioned": "2026-05-22T10:00:00Z",
  "evidence_snippet": "用户：昨天那篇论文的实验部分…"
}
```

---

## 7. Host 注入格式（建议）

Host 将 MCP 返回拼入 `planInput`（顺序 **先于** Memory hints）：

```markdown
## 协作人格（Soul）
{persona_prompt}

## 近期议题与事件（Soul）
{event_context}
```

分隔符与标题以 `soul.config.prompt_blocks` 为准；变更须 bump `version`。

---

## 8. 可观测性

| 信号 | 用途 |
|------|------|
| `retrieve_meta.degraded` | Host 决定是否提示用户「记忆暂不可用」 |
| store queue depth | 运维背压 |
| episode 归档 | 人工审计抽取质量 |

Companion Web 控制台：**非 MVP**；可后续对齐 Memory MCP `:8091` 模式。

---

## 9. 安全与隐私

- `data/` 含用户对话摘录：**不得**提交进公开 git（`.gitignore`）。  
- `user_id` 隔离目录（实现时 `data/users/{id}/`）。  
- 无密钥写入 `soul.config`；LLM API key 仅 `mcp_env`。

---

## 10. 测试策略（实现后）

| 层级 | 内容 |
|------|------|
| 单元 | template 组装、profile merge、event upsert |
| MCP 契约 | store ACK + async job；retrieve 字段禁止 Memory 键 |
| Host 集成 | 两轮 WebUI 场景（见 `ACCEPTANCE_RULES.md`） |

---

## 11. 版本与兼容

- 工具名冻结前可使用 `phase=0-docs-only`。  
- JSON 字段增删须 bump `retrieve_meta.schema_version`。  
- Host 与 MCP 版本协商：环境变量 `SOUL_MCP_MIN_HOST`（待定）。
