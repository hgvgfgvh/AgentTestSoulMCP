# Soul MCP — 验收规则

> 状态：**规划**（全部未实现）。实现后勾选 `IMPLEMENTATION_PROGRESS.md`。

---

## A. 工具面（MCP）

| ID | 规则 | 验证方式 |
|----|------|----------|
| A1 | 仅注册 `soul_store`、`soul_retrieve` | MCP `tools/list` 快照 |
| A2 | 入参/出参为 **字符串 JSON** | 契约测试 |
| A3 | `soul_retrieve` 响应 **无** `exec_simple_match` / Memory 路由键 | JSON schema 测试 |
| A4 | `soul_store` 同步 &lt; 200ms 返回 `accepted`（不含 async 整理） | 压测单测 |

---

## B. Store 管道

| ID | 规则 | 验证方式 |
|----|------|----------|
| B1 | 合法 `context` 入队后 `episodes/{id}.json` 存在 | 集成测试 |
| B2 | 无 LLM 时 rules 仍可写入至少 1 条 event 或 profile | fixture 对话 |
| B3 | 有 LLM 时失败降级 rules，不丢 episode 原文 | mock LLM down |
| B4 | **不**自动改写 `soul.config` | 文件 mtime 断言 |
| B5 | overlay 仅经显式路径写入 `soul_overlay/` | 审计目录列表 |

---

## C. Retrieve 管道

| ID | 规则 | 验证方式 |
|----|------|----------|
| C1 | 默认无 LLM 仍返回非空 `persona_prompt`（含 config 基座） | 空 data 目录测试 |
| C2 | 预置 events 后，用户短问句可命中 `event_context` | 「昨天那篇论文」fixture |
| C3 | 超预算截断 `event_context`，`retrieve_meta` 标明 | token 计数单测 |
| C4 | 超时返回 `degraded: true`，Host 可继续 | 注入慢 LLM mock |
| C5 | `persona_prompt` 含 profile 中 `preferred_name` 类条目 | 两轮对话集成 |

---

## D. 与 Memory 隔离

| ID | 规则 | 验证方式 |
|----|------|----------|
| D1 | Soul store **不**调用 Memory MCP | 进程边界 / 无 memory 工具注册 |
| D2 | Soul 文档与 Memory 文档互引边界一致 | 人工评审清单 |
| D3 | 同一段口头禅 **不**出现在 Memory facts 主路径（Host 配置下 E2E） | 双 MCP 集成测试 |

---

## E. Host 集成（AgentTest）

| ID | 规则 | 验证方式 |
|----|------|----------|
| E1 | `soul_retrieve` 在 `memory_retrieve` **之前** | gateway 单测 / 日志顺序 |
| E2 | Store 源为 WebUI 序列化，非 Behavior tool 日志 | mock 序列化器 |
| E3 | Soul 宕机时 Plan 仍完成一轮 | 杀 soul-mcp 进程测试 |
| E4 | `plan_soul_hook.enabled: false` 时零调用 | 配置测试 |

主项目对应：`AgentTest/.../ACCEPTANCE_RULES.md` §J。

---

## F. 场景验收（人工 / E2E）

**场景 F-1 — 跨会话议题**

1. 会话 A：用户与 Agent 讨论「项目 X」「论文 Y」≥3 轮；回合结束触发 store。  
2. 会话 B（新 session）：用户仅输入「昨天论文 Y 的结论是什么？」  
3. **通过**：`event_context` 含 Y；Plan 回复不索要全文背景。

**场景 F-2 — 口头禅 / 称呼**

1. 用户说「叫我老王」「少说废话」。  
2. 下一轮 store 完成后新 retrieve。  
3. **通过**：`persona_prompt` 可见对应 profile；回复长度变短（人工抽检）。

**场景 F-3 — 降级**

1. 停止 soul-mcp。  
2. 用户正常提问。  
3. **通过**：门户有回复；无 panic；可选日志 `soul hook degraded`。

---

## 验收命令（实现后填入）

```bash
# Soul MCP 仓库
go test ./...

# Host（集成，待 soulhook 存在）
go test ./plan/soulhook/...
```
