# Soul MCP — 设计意图（宪法）

> **维护规则**：人工维护；Agent 不得用实现便利覆盖下文。新决策以日期段落追加。  
> **宿主约束**：`AgentTest` 阶段三铁律 F3-1～F3-9 优先约束 **Host 如何调用**；本文约束 **Soul MCP 内部**。

---

## 1. 要解决什么问题

| 痛点 | 意图 |
|------|------|
| 跨日/跨会话重复铺垫（项目、论文、议题） | **历史事件配置** + retrieve 指代，实现「一点就通」 |
| 用户口头禅、称呼、回复长度偏好 | **用户人格画像**（可验证条目），协作风格同步 |
| 进入系统冷启动感 | 首轮 `soul_retrieve` 即注入 persona + 议题，无需用户复述 |
| 执行仍要快、要稳 | **不**替代 Memory MCP；不输出 Exec-Simple 路由 |

本系统是 **自我进化体系的前期投入**：先建立 **可审计、可钩子化、进程隔离** 的「人—事—风格」层；**不包含**自主修改 Host 代码或宪法。

---

## 2. 铁律（S1～S9）

| # | 铁律 | 含义 |
|---|------|------|
| S1 | **双工具面** | 对外 **仅** `soul_store`、`soul_retrieve`；字符串 JSON 入参/出参 |
| S2 | **存慢取快** | `soul_store`：同步 ACK + 队列异步整理（可含 LLM）；`soul_retrieve`：同步、有超时与模板回退 |
| S3 | **WebUI 对话为源** | Store 只接受 Host 序列化的 **用户↔主 Agent WebUI** 对话；拒绝 Behavior 步内 tool 轨迹作为主源 |
| S4 | **与 Memory 正交** | 禁止把人格/口头禅/议题作为主事实写入 Memory `facts.jsonl`；禁止输出 `exec_simple_match` |
| S5 | **soul.config 基座只读** | 进程旁 `soul.config` 为稳定人格；LLM 维护内容落 `data/soul_overlay/`，可审计、可回滚 |
| S6 | **retrieve 默认无 LLM** | MVP 用模板组装 `persona_prompt` + `event_context`；可选 `SOUL_RETRIEVE_LLM=1` 须预算与超时 |
| S7 | **失败可降级** | store 失败记日志；retrieve 超时返回空块 + `degraded: true`，不抛致命错误给 Host |
| S8 | **多租户预留** | `user_id` / `session_id` 字段版本化；未传时单用户默认 |
| S9 | **进化边界** | 可自动整理 profile/events；**不可**自动改 `soul.config`、不可改 Host 配置、不可调 Memory 工具 |

---

## 3. 核心概念

### 3.1 用户人格画像（profile）

从对话中抽取 **稳定、可验证** 的用户特征，例如：

- 称呼偏好（「叫我老王」）
- 口头禅 / 高频短语
- 回复长度、正式程度
- 领域兴趣（非执行路径）

存储形态：`data/profile.jsonl`（一行一条，带 `confidence`、`source_episode_id`、`updated_at`）。

### 3.2 历史事件 / 议题配置（events）

从对话中抽取 **可指代** 的议题节点，例如：

- 项目名、论文标题、会议、待办主题
- `last_mentioned`、关联实体、一句话摘要
- 与用户原话的可追溯 `evidence_snippet`（短摘录）

存储形态：`data/events.jsonl` + 可选 `data/episodes/` 原始归档。

### 3.3 soul.config（Agent 人格基座）

与 MCP 可执行文件 **同级** 的 YAML：

- `persona.role` / `tone` / `boundaries`
- `prompt_blocks` 标题
- retrieve 时 **只读合并** 进 `persona_prompt`

动态「进化」若需调整口吻，写 **overlay**，不写回基座（S5）。

### 3.4 soul_overlay（可选）

`data/soul_overlay/active.yml`：经 LLM 或人工审核的 **增量人格补丁**（例如「本周用户强调论文写作模式」）。  
须带 `reason`、`created_at`；retrieve 时叠加在基座之后、profile 之前或之后（实现冻结一种顺序）。

---

## 4. 工具契约（草案 · 实现前须 Host 联签）

### 4.1 soul_store

**Host 调用时机**：WebUI 回合结束（及可选：长对话分段，待冻结）。

**入参（JSON 字符串）**：

```json
{
  "context": "【必需】Host 序列化的本轮 WebUI 对话全文或增量",
  "user_id": "可选，默认 default",
  "session_id": "可选",
  "episode_id": "可选，Host 生成 UUID",
  "metadata": { "channel": "webui", "locale": "zh-CN" }
}
```

**同步返回**：

```json
{
  "accepted": true,
  "episode_id": "...",
  "queue_depth": 0
}
```

**异步（MCP 内部）**：LLM/规则抽取 → 合并 profile → 追加/更新 events → 可选写 overlay 建议（**不**自动应用 overlay）。

### 4.2 soul_retrieve

**Host 调用时机**：用户新消息进入 Plan **之前**（早于 `memory_retrieve`）。

**入参**：

```json
{
  "context": "【必需】当前用户输入 + 可选本轮迄今 WebUI 摘录",
  "user_id": "可选",
  "session_id": "可选",
  "budget_tokens": 1500
}
```

**出参**：

```json
{
  "persona_prompt": "markdown 块，含 soul.config + profile + overlay",
  "event_context": "markdown 块，议题列表与指代提示",
  "retrieve_meta": {
    "profile_hits": 3,
    "event_hits": 2,
    "degraded": false,
    "latency_ms": 12
  }
}
```

**禁止字段**：`exec_simple_match`、`memory_hints`、`tool_chain` 等 Memory 专属键（S4）。

---

## 5. 与 Memory MCP 的边界

| 维度 | Soul MCP | Memory MCP |
|------|----------|------------|
| 典型问题 | 「我们昨天聊的那篇论文」 | 「上次用 PowerShell 装依赖失败了」 |
| 存储内容 | profile、events、episodes | facts、graph、pitfall |
| 影响 Plan 路由 | **否** | **是**（Exec-Simple） |
| LLM 用法 | store 整理、可选 retrieve compose | fact 抽取、L2 冲突、可选 2e prune |

Host 须对同一 `context` 文本做 **去重策略**（避免双 MCP 各跑一遍相同 RAG），细则在 Host `ARCHITECTURE` §15 冻结。

---

## 6. 非目标（v1）

- 不提供控制台以外的 HTTP API（Companion 控制台可后续单独立项）。  
- 不替代 `sessionmemory` 的近轮原话缓存。  
- 不生成 SKILL / 不改 `skill_packs`。  
- 不做全自动「改 soul.config」；overlay 须可审计。  
- 不承担 Affective 情绪模型的科学仿真（仅协作口吻适配）。

---

## 7. 决策日志

### 2026-05-24 — 仓库与钩子模式确立

- **决策**：独立 repo `AgentTestSoulMCP` + Host `plan_soul_hook`，对称 Memory MCP。  
- **原因**：舒适度与执行经验正交；进程隔离便于演进与降级。  
- **状态**：仅文档，无代码。

**开放问题**：

1. retrieve 是否默认附带最近 N 条 profile 全量 vs 仅命中  
2. events 过期策略（TTL vs 显式 supersede）  
3. 是否与 `AgentTest/agent/soul/Nexus.yml` 合并或替代  
4. 多用户 `user_id` 来源（WebUI 登录态）
