# Soul MCP — 架构漂移登记

对比 `DESIGN_INTENT.md`（意图）与仓库代码（实现）。  
Agent **不得**将「违反设计」项自动合法化进 `ARCHITECTURE.md`。

---

## 当前总览（2026-05-24）

| 分类 | 数量 | 说明 |
|------|------|------|
| 技术债 | 全部 | **无 Go 代码** |
| 符合意图 | 0 | 无实现可对齐 |
| 违反设计 | 0 | — |

---

## 登记项

### SD-1 双工具 MCP — 技术债

| 项 | 状态 | 说明 |
|----|------|------|
| `soul_store` / `soul_retrieve` | **技术债** | 未实现 |
| 字符串 JSON 契约 | **技术债** | 仅文档 §4 |

---

### SD-2 数据目录 — 技术债

| 项 | 状态 | 说明 |
|----|------|------|
| `profile.jsonl` / `events.jsonl` | **技术债** | 未创建运行时目录 |
| `episodes/` 归档 | **技术债** | 未实现 |
| `soul_overlay/` | **技术债** | 未实现 |

---

### SD-3 soul.config — 技术债

| 项 | 状态 | 说明 |
|----|------|------|
| 基座只读加载 | **技术债** | 仅有 `soul.config.example` |
| overlay 审计 | **技术债** | 未实现 |

---

### SD-4 Host 钩子 — 技术债（跨仓库）

| 项 | 状态 | 说明 |
|----|------|------|
| `plan_soul_hook` | **技术债** | `AgentTest` 未实现 |
| 注入顺序 soul→memory | **技术债** | `portal/gateway` 仅 Memory |

---

### SD-5 retrieve 默认无 LLM — 符合意图（规划）

| 项 | 状态 | 说明 |
|----|------|------|
| 模板组装默认 | **符合意图** | 宪法 S6；待实现验证 |

---

### SD-6 Nexus.yml 与 Soul — 需要人工决策

| 项 | 状态 | 说明 |
|----|------|------|
| `AgentTest/agent/soul/Nexus.yml` vs Soul retrieve | **需要人工决策** | Affective 双源人格是否合并 |

---

## 优先级（人工）

1. **P0**：冻结 JSON 契约 + `soul-mcp` skeleton + 模板 retrieve  
2. **P1**：store 队列 + episode 归档 + rules 抽取  
3. **P2**：Host `plan/soulhook` + portal 顺序  
4. **P3**：可选 store/retrieve LLM + overlay 工作流  

---

## 变更记录

| 日期 | 变更 |
|------|------|
| 2026-05-24 | 初版：文档-only 仓库，全部 SD 为技术债 |
