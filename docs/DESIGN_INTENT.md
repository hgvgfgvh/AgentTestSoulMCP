# Soul MCP — 设计意图（宪法）

> **维护规则**：人工维护；Agent 不得用实现便利覆盖下文。新决策以日期段落追加。  
> **宿主约束**：`AgentTest` 阶段三铁律 F3-1～F3-9 约束 **Host 如何调用**；本文约束 **Soul MCP 进程内部**（主项目无需了解 MCP 实现细节）。

---

## 1. 要解决什么问题

| 痛点 | 意图 |
|------|------|
| 跨日/跨会话重复铺垫（项目、论文、议题） | **按天落地事实** + **地图热索引** + retrieve 指代，实现「一点就通」 |
| 用户口头禅、称呼、思维链路、情绪与人格归纳 | **person.md**（LLM 按需维护的可读画像） |
| 进入系统冷启动感 | 首轮 `soul_retrieve` 注入 **预取缓存 + 地图 + 画像 + Agent 灵魂** |
| 长上下文注意力涣散 | **存入分治法**：禁止单次 LLM 包办 store；**取出快慢双轨** |
| 执行仍要快、要稳 | **不**替代 Memory MCP；**不**输出 Exec-Simple / Memory 路由 |

本系统是 **自我进化体系的前期投入**：建立可审计、可钩子化、进程隔离的「人—事—风格—时空因果」层；**不包含**自主修改 Host 代码或主项目宪法。

---

## 2. 铁律（S1～S10）

| # | 铁律 | 含义 |
|---|------|------|
| S1 | **双工具面** | 对外 **仅** `soul_store`、`soul_retrieve`；字符串 JSON 入参/出参 |
| S2 | **存慢取快** | `soul_store`：同步 ACK + **异步**多路整理；`soul_retrieve`：同步、可超时、有模板回退 |
| S3 | **WebUI 对话为源** | Store 只接受 Host 序列化的 **用户↔主 Agent WebUI** 对话；不以 Behavior 步内 tool 轨迹作为主源 |
| S4 | **与 Memory 正交** | 禁止把人格/口头禅/议题作为主事实写入 Memory `facts.jsonl`；禁止输出 `exec_simple_match` 等 Memory 键 |
| S5 | **Agent 灵魂只读** | `soul.agent.yaml`（或 `SOUL_MCP_SOUL_DOC`）为用户/运维定义；**MCP 内 LLM 不得改写**；retrieve 最终 hints **必须**体现其基调 |
| S6 | **存入职责分离** | Store **禁止**单次 LLM 同时写落地日文件、画像、地图、预取；须 **≥4 次独立 LLM 调用**（或等价独立任务） |
| S7 | **失败可降级** | store 任一路失败记日志、不阻断 ACK；retrieve 失败返回空 `hints` 或模板块，不抛致命错误给 Host |
| S8 | **多租户预留** | `user_id` / `session_id` 字段版本化；未传时单用户默认 `data/` |
| S9 | **进化边界** | 可自动整理 **person / map / 按天事实 / llm_cache**；**不可**自动改 `soul.agent.yaml`、不可改 Host 配置、不可调 Memory 工具 |
| S10 | **历史事实四维标签** | 按天 JSONL 每条须使用 **现象 / 时空 / 因果 / 存在** 四维结构（见 §3.1）；禁止退回扁平 `tags-only` 作为主形态 |

---

## 3. 核心概念

### 3.1 历史事实 — 哲学四维标签（按天冷存储）

**目的**：对齐「实体与范畴 / 先验时空 / 因果冲动 / 此在与共在」的划分，使存入可索引、取出可检索。

| 维度 | JSON 键 | 子字段 | 含义摘要 |
|------|--------|--------|----------|
| **现象与客体** | `phenomenon` | `entity`, `category`, `artifacts` | 指向物、知识范畴、文件/配置留痕 |
| **先验时空** | `spatiotemporal` | `chronos`, `kairos`, `domain` | ISO 时序、项目阶段契机、逻辑作用域 |
| **行为因果** | `causality` | `intent`, `action`, `outcome`, `evolution_potential` | 动机、手段、终态、是否值得沉淀 Skill |
| **存在与互涉** | `existential` | `cognitive_align`, `mood_tone`, `persona_shift` | 默契度、情绪底色、激活的人格投影 |

**存储形态**：`{SOUL_MCP_DATA_DIR}/storage/history/YYYY-MM-DD.jsonl`（一行一条 JSON）。

**通用字段**（非第五维）：`id`, `summary`, `evidence`, `stored_at`, `source`, `relations`。

声明式枚举与 LLM 说明：`agentConfig/soul-agent.yaml` → `fact_dimensions` + `llm.store_daily_system`。

### 3.2 用户画像（person.md）

从对话中归纳 **稳定、可协作** 的用户特征，Markdown 自由章节，例如：

- 称呼与身份、习惯与偏好、常用话术
- 思维与解决问题方式、情绪与互动风格
- 其它观察（含用户自述之「23 形态」等人格归纳时 **须标注不确定性**）

**维护**：store **任务2** 专用 LLM，**整文件覆盖**（带 `.bak`）。  
**不是**四维标签表；与 `existential.persona_shift`（单条事实级）互补。

### 3.3 地图文档（map.md）— 热索引

**目的**：最近几天 **高价值摘要 + 标签 + 落地指针**，避免全库扫描。

每条须含：

- 标题式摘要、梳理好的标签（Entity / Category / Kairos 等）
- 指针：`storage/history/YYYY-MM-DD.jsonl` 或 `fact-id`

**维护**：store **任务3** 专用 LLM；不记录代码细节。

### 3.4 LLM 级预取缓存（llm_cache.json）

**目的**：store 末尾根据 **最近对话 + person + map** 预测用户后续问题（默认最多 5 个，可配置），对每个问题从 **地图 + 按天文件** 检索材料，写入缓存供 **下一轮 retrieve 快通道** 使用。

**维护**：store **任务4**（须在任务 1～3 完成后执行，以便指针与画像最新）。

### 3.5 Agent 灵魂（soul.agent.yaml）

与 MCP 可执行文件同级或通过 `SOUL_MCP_SOUL_DOC` 指定：

- 用户定义的 Agent 口吻、边界、角色
- retrieve 时 **只读**；最终 `hints` 顶层含「Agent 灵魂」节（S5）

**禁止**：任何 store 任务写入或覆盖该文件。

### 3.6 与已废弃形态的关系

| 旧形态 | 状态 |
|--------|------|
| `profile.jsonl` / `events.jsonl` 主路径 | **已废弃**；若 `data/` 残留仅作迁移参考 |
| `soul.config` + `soul_overlay/` 主路径 | **已废弃**；由 `soul.agent.yaml` + `person.md` 替代 |
| 单文件 `history.facts.jsonl` | **兼容只读**；retrieve 可与按天文件合并检索 |

---

## 4. 工具契约（Host 联调冻结）

### 4.1 soul_store

**Host 调用时机**：WebUI 回合结束（`plan_soul_hook` 异步，不阻塞主回复）。

**入参（工具 arguments）**：

| 字段 | 说明 |
|------|------|
| `content` | **必需**。Host 序列化的本轮 WebUI 对话 |
| `source` | 可选，如 `agenttest-webui` |
| `kind` | 可选 |
| `correlation_id` | 可选，用于 `job_id` |

**同步返回（JSON 字符串）**：

```json
{
  "accepted": "true",
  "job_id": "soul-job-…",
  "skipped": "false",
  "message": "accepted; async store (daily+person+map+prefetch)",
  "phase": "4-async-pipeline"
}
```

**异步（MCP 内部，S6）**：

1. 任务1：LLM → 按天 JSONL + 四维标签  
2. 任务2：LLM → `person.md`  
3. 任务3：LLM → `map.md`  
4. 任务4：LLM 预测问题 → Go `recall` 检索 → `llm_cache.json`  

### 4.2 soul_retrieve

**Host 调用时机**：用户新消息进入 Plan **之前**（早于 `memory_retrieve`）。

**入参**：

| 字段 | 说明 |
|------|------|
| `context` | **必需**。含「用户输入: …」等上下文 |
| `query_hint` | 可选，优先作检索 query |

**出参（JSON 字符串）**：

```json
{
  "hints": "## Agent 灵魂（用户定义·只读）\n…\n\n## Soul 协作提示\n…",
  "skipped": "false",
  "phase": "4-async-pipeline"
}
```

Host 从 JSON 取 **`hints`** 单字段拼入 `planInput`（见 `ARCHITECTURE.md` §7）。  
**禁止**在 Soul 响应中带 Memory 路由键（S4）。

**内部快慢双轨**（Host 不可见）：

- **快通道**：`llm_cache` + `map` + `person` 足够 → Gate LLM 一次总结  
- **慢通道**：Gate 输出 `retrieval_tags` → Go 检索按天文件 → Compose LLM 二次总结  

---

## 5. 与 Memory MCP 的边界

| 维度 | Soul MCP | Memory MCP |
|------|----------|------------|
| 典型问题 | 「我们昨天聊的那篇论文」「项目卡点」 | 「上次 PowerShell 装依赖失败」 |
| 存储 | 按天四维事实、person、map、cache | facts、graph、pitfall |
| 影响 Plan 路由 | **否** | **是**（Exec-Simple） |
| LLM | 多任务 store；1～2 次 retrieve | fact 抽取、冲突、prune |

---

## 6. 非目标

- 不提供 Soul 专属 HTTP 控制台（非 MVP）。  
- 不替代 Host 会话近轮 buffer。  
- 不生成 SKILL / 不改 `skill_packs`。  
- 不自动修改 `soul.agent.yaml`。  
- 不承担临床级情绪/人格科学仿真。  
- 不在本文档规定 Host 门户、Plan 拆步逻辑（属 `AgentTest`）。

---

## 7. 决策日志

### 2026-05-24 — 仓库与钩子模式确立

- 独立 repo + Host `plan_soul_hook`，与 Memory 对称。  
- **状态**：已被 v4 取代主数据形态，钩子契约保留。

### 2026-05-25 — v4 分治法与四维标签冻结

- **决策**：按天 `storage/history/` + `person.md` + `map.md` + `llm_cache.json`；store 四路异步 LLM；retrieve 快慢双轨；历史事实采用哲学 **四维 JSON**（§3.1）。  
- **原因**：避免单次 store 上下文污染；地图作热索引；预取降低延迟；四维对齐用户「哲学世界」标签体系。  
- **Agent 灵魂**：仅 `soul.agent.yaml` 只读，不再以 `soul.config` 为 MCP 内基座。  
- **对外契约**：仍仅 `soul_store` / `soul_retrieve`，`hints` 单字段返回。

**开放问题**：

1. `map` 指针的 Go 侧强校验（格式 / 存在性）  
2. 多用户 `data/users/{id}/` 落地时间表  
3. 快通道 Gate 误判率与是否默认慢通道  
4. 向量检索是否作为第五通道接入 `internal/recall`
