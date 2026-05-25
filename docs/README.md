# Soul MCP 文档集

本目录为 **AgentTestSoulMCP** 的设计 SSOT（**仅本仓库**；主项目 `AgentTest` 不承载 MCP 内部实现说明）。

## 阅读顺序

1. **`DESIGN_INTENT.md`** — 宪法：铁律 S1～S10、哲学四维标签、工具契约、与 Memory 边界  
2. **`ARCHITECTURE.md`** — v4 实现：四路 store、快慢 retrieve、目录布局、模块表  
3. **`ACCEPTANCE_RULES.md`** — 验收条目与场景  
4. **`ARCHITECTURE_DRIFT.md`** — 意图 vs 代码漂移登记  
5. **`IMPLEMENTATION_PROGRESS.md`** — 里程碑勾选  

## 当前版本摘要

| 项 | 值 |
|----|-----|
| phase | `4-async-pipeline` |
| 工具 | `soul_store`, `soul_retrieve` |
| 冷存储 | `storage/history/YYYY-MM-DD.jsonl`（四维标签） |
| 热索引 | `map.md` |
| 画像 | `person.md` |
| 预取 | `llm_cache.json` |
| 只读灵魂 | `soul.agent.yaml` |

## 与主项目关系

| 仓库 | 职责 |
|------|------|
| **AgentTest** | WebUI、`plan_soul_hook` 调用 MCP；只解析 `hints` JSON 字段 |
| **AgentTestSoulMCP**（本仓库） | 全部记忆整理与检索逻辑 |
| **AgentTestMemoryMCP** | 执行经验；与 Soul 正交 |

主项目仅需：`config/app.yaml` 中 `plan_soul_hook` 的 `mcp_command` / `mcp_env`，**无需**同步本目录文档。

## 仓库根目录

- `README.md` — 部署与 env 速查  
- `agentConfig/soul-agent.yaml` — 运行时规则与 LLM 提示词  
