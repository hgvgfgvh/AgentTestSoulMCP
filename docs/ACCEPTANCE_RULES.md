# Soul MCP — 验收规则

> **对齐实现**：v4 `4-async-pipeline`（2026-05-25）。勾选状态见 `IMPLEMENTATION_PROGRESS.md`。

---

## A. 工具面（MCP）

| ID | 规则 | 验证方式 | 状态 |
|----|------|----------|------|
| A1 | 仅注册 `soul_store`、`soul_retrieve` | `tools/list` | ✅ |
| A2 | 入参/出参为字符串 JSON | 契约测试 | ✅ |
| A3 | `soul_retrieve` 响应 **无** Memory 路由键；含 `hints` | JSON 解析 | ✅ |
| A4 | `soul_store` 同步快速返回 `accepted` + `job_id` | 单测/边界测试 | ✅ |
| A5 | `phase` 为 `4-async-pipeline`（或兼容旧 phase） | 响应字段 | ✅ |

---

## B. Store 管道（四路异步）

| ID | 规则 | 验证方式 | 状态 |
|----|------|----------|------|
| B1 | 有 LLM 时任务1 写入 `storage/history/YYYY-MM-DD.jsonl` | 文件存在且 JSONL 合法 | ✅ 需 LLM |
| B2 | 任务2 可更新 `person.md`（带 `.bak`） | store 后 mtime/内容 | ✅ 需 LLM |
| B3 | 任务3 可更新 `map.md` | 同上 | ✅ 需 LLM |
| B4 | 任务4 可写入 `llm_cache.json`（含 predicted_questions） | 文件 JSON | ✅ 需 LLM |
| B5 | 四路为 **独立** LLM 提示（非单 prompt 合并） | 代码审查 `store_tasks.go` | ✅ |
| B6 | 无 LLM 时降级：仅今日文件一条事实 | 无 API 环境测试 | ✅ |
| B7 | **不**改写 `soul.agent.yaml` | mtime/只读断言 | ✅ |
| B8 | store 失败不导致同步 ACK 失败 | 注入 LLM 错误 | ✅ |

---

## C. 历史事实 — 四维标签

| ID | 规则 | 验证方式 | 状态 |
|----|------|----------|------|
| C1 | 每条含 `phenomenon` / `spatiotemporal` / `causality` / `existential` 对象 | JSON schema 抽检 | ⚠️ 依赖 LLM |
| C2 | `entity` / `category` / `artifacts` 语义符合设计 | 人工/黄金对话 | ⬜ |
| C3 | `chronos` 空时由系统补 `stored_at` | `AppendDay` 单测 | ✅ |
| C4 | 旧版 `tags`/`entities` 只读兼容 `NormalizeLegacy` | 单测 | ✅ |

---

## D. Retrieve 管道（快慢双轨）

| ID | 规则 | 验证方式 | 状态 |
|----|------|----------|------|
| D1 | 无 LLM 时仍返回非空 `hints`（含 soul + 模板） | engine 单测 | ✅ |
| D2 | 有 LLM 时 Gate 可走快通道或吐 `retrieval_tags` | mock/integration | ⚠️ 需 LLM |
| D3 | 慢通道经 `recall.Select` 再 Compose | 代码路径 | ✅ |
| D4 | 最终 `hints` 含 **Agent 灵魂** 节（只读摘录） | 字符串断言 | ✅ |
| D5 | 问「昨天」类问题可命中时间窗（有昨日数据） | `recall` 单测 | ✅ |
| D6 | 超 `max_hints_runes` 截断 | 配置 + FormatFinalHints | ✅ |

---

## E. 与 Memory 隔离

| ID | 规则 | 验证方式 | 状态 |
|----|------|----------|------|
| E1 | Soul 进程 **无** Memory 工具 | MCP list | ✅ |
| E2 | hints **无** `exec_simple_match` | 边界测试 M05 | ✅ |
| E3 | 人格/议题不写 Memory facts（Host E2E） | 主项目双 MCP 测试 | ⬜ |

---

## F. Host 集成（AgentTest，契约级）

| ID | 规则 | 验证方式 | 状态 |
|----|------|----------|------|
| F1 | `soul_retrieve` 在 `memory_retrieve` **之前** | gateway 日志 | ✅ |
| F2 | Store 材料为 WebUI 对话 | `BuildWebUIDialogueContent` | ✅ |
| F3 | Soul 失败时 Plan 仍继续（空 hints） | 杀进程测试 | ✅ |
| F4 | `plan_soul_hook.enabled: false` 零调用 | 配置测试 | ✅ |
| F5 | Host 只消费 `hints` 字段 | `extractHintsField` | ✅ |

---

## G. 场景验收（人工 / E2E）

**G-1 — 跨会话议题**

1. 会话 A：讨论项目/论文 ≥3 轮 → store。  
2. 会话 B：「昨天论文 Y 的结论？」  
3. **通过**：hints 含 Y 相关事实或地图指针；Plan 不索要全文背景。

**G-2 — 称呼 / 习惯**

1. 用户「叫我老王」→ store。  
2. 下一轮 retrieve。  
3. **通过**：person 或 hints 体现称呼。

**G-3 — 预取快通道**

1. store 完成且 `llm_cache` 非空。  
2. 用户问与 predicted_questions 相近的一句。  
3. **通过**：Gate 倾向 sufficient（人工抽检）。

**G-4 — 降级**

1. 停止 soul-mcp 或清空 LLM env。  
2. 用户提问。  
3. **通过**：门户有回复；hints 为空或模板。

---

## 验收命令

```bash
# 本仓库
cd AgentTestSoulMCP
go test ./...

# Host 边界（主项目根目录）
go run ./scripts/soul_boundary_test
```
