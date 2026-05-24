# Soul MCP 文档集

本目录为 **AgentTestSoulMCP** 的设计 SSOT。编码前请先读 `DESIGN_INTENT.md`，再读 `ARCHITECTURE.md` 与 `ACCEPTANCE_RULES.md`。

## 阅读顺序

1. `DESIGN_INTENT.md` — 不可破坏的边界（铁律 S1～S9）
2. `ARCHITECTURE.md` — store/retrieve、数据目录、`soul.config`
3. `ACCEPTANCE_RULES.md` — MVP 与集成验收
4. `ARCHITECTURE_DRIFT.md` — 规划期全部为「未实现」
5. `IMPLEMENTATION_PROGRESS.md` — 里程碑勾选

## 与主项目关系

| 仓库 | 职责 |
|------|------|
| `AgentTest` | WebUI、Plan/Exec、`plan_soul_hook` 序列化 WebUI 对话并注入 hints |
| `AgentTestSoulMCP`（本仓库） | `soul_store` / `soul_retrieve`、profile/events 整理 |
| `AgentTestMemoryMCP` | `memory_store` / `memory_retrieve`；**禁止**写入人格/口头禅主路径 |

主项目阶段三宪法：`AgentTest/Agent编码防止架构坍塌的处理方法论/DESIGN_INTENT.md`（F3-1～F3-9）。
