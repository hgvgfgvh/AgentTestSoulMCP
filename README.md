# AgentTestSoulMCP

外置 **Soul MCP**（人格 / 用户画像 / 议题续接 / 协作适配）服务，与主项目 `AgentTest` 通过 **Host 钩子**（规划：`plan_soul_hook`）集成。

与 `AgentTestMemoryMCP` **同级、解耦**：Memory 管「怎么执行」；Soul 管「在聊什么、用户是谁、怎么说话」。

## 状态

| 项 | 状态 |
|----|------|
| 设计/架构/验收文档 | **初版**（2026-05-24） |
| Go 实现 / `soul-mcp` 可执行文件 | **未开始** |
| Host `plan_soul_hook` | **未开始**（主项目） |

## 文档

| 文件 | 用途 |
|------|------|
| [docs/DESIGN_INTENT.md](./docs/DESIGN_INTENT.md) | 子系统宪法 |
| [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | 模块与数据流 |
| [docs/ACCEPTANCE_RULES.md](./docs/ACCEPTANCE_RULES.md) | 可执行验收 |
| [docs/ARCHITECTURE_DRIFT.md](./docs/ARCHITECTURE_DRIFT.md) | 意图 vs 实现差异 |
| [docs/IMPLEMENTATION_PROGRESS.md](./docs/IMPLEMENTATION_PROGRESS.md) | 实现进度 |

## 配置样例

- [soul.config.example](./soul.config.example) — Agent **人格基座**（MCP retrieve 时合并；默认只读）

## 主项目交叉引用

- Host 宪法：`AgentTest/Agent编码防止架构坍塌的处理方法论/DESIGN_INTENT.md`（阶段三 F3-1～F3-9）
- Host 目标架构：`.../ARCHITECTURE.md` §15
- Memory 边界：`AgentTestMemoryMCP/docs/DESIGN_INTENT.md`（不承载人格/口头禅）
