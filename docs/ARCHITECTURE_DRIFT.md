# Soul MCP — 架构漂移登记

对比 `DESIGN_INTENT.md`（意图）与仓库代码（实现）。  
Agent **不得**将「违反设计」项自动合法化进 `ARCHITECTURE.md`。

---

## 当前总览（2026-05-25 · v4）

| 分类 | 数量 | 说明 |
|------|------|------|
| 符合意图 | 多数核心 | 双工具、四路 store、双轨 retrieve、按天四维、soul 只读 |
| 技术债 | 若干 | 见下表 SD-* |
| 违反设计 | 0 | — |
| 遗留数据 | 1 | `data/` 下旧 profile/events 文件未自动迁移 |

---

## 登记项

### SD-1 双工具 MCP — 符合意图

| 项 | 状态 |
|----|------|
| `soul_store` / `soul_retrieve` | ✅ 已实现 |
| 字符串 JSON；retrieve 返回 `hints` | ✅ |

---

### SD-2 数据布局 v4 — 符合意图

| 项 | 状态 |
|----|------|
| `storage/history/YYYY-MM-DD.jsonl` | ✅ |
| `person.md` / `map.md` / `llm_cache.json` | ✅ |
| 四维 `Fact` 结构 | ✅ `fact_dims.go` |

---

### SD-3 Agent 灵魂 — 符合意图

| 项 | 状态 |
|----|------|
| `soul.agent.yaml` 只读 | ✅ 无 store 写入路径 |
| 最终 hints 含灵魂节 | ✅ `FormatFinalHints` |

---

### SD-4 存入分治 — 符合意图

| 项 | 状态 |
|----|------|
| 四路独立 LLM | ✅ `store_tasks.go` + `store_pipeline.go` |
| 预取在 wave2 | ✅ 任务1～3 完成后 |

---

### SD-5 retrieve — 部分技术债

| 项 | 状态 | 说明 |
|----|------|------|
| 快慢双轨 | ✅ | Gate + Compose |
| `retrieval_tags` 非四维 | **符合意图** | 故意扁平，供 Go 检索 |
| Gate 误判率未度量 | **技术债** | 无黄金集 |
| map 指针未 Go 校验 | **技术债** | 仅 LLM 维护 |

---

### SD-6 废弃路径 — 遗留/漂移

| 项 | 状态 | 说明 |
|----|------|------|
| `profile.jsonl` / `events.jsonl` 主存储 | **已废弃** | 代码不再写入；`data/` 可能残留 |
| `soul.config` + overlay | **已废弃** | 由 `soul.agent.yaml` + `person.md` 替代 |
| `history.facts.jsonl` 单文件 | **兼容** | 只读合并检索 |
| `internal/profile` `internal/events` 包 | **技术债** | 死代码，待删或标 deprecated |

---

### SD-7 Host 钩子 — 符合意图（跨仓库）

| 项 | 状态 |
|----|------|
| `plan_soul_hook` | ✅ AgentTest |
| soul → memory 顺序 | ✅ gateway |

---

### SD-8 多租户 — 技术债

| 项 | 状态 |
|----|------|
| `data/users/{id}/` | ⬜ 未实现 |

---

### SD-9 向量检索 — 技术债

| 项 | 状态 |
|----|------|
| embedding 通道 | ⬜ 仅词重合 + 时间窗 + RRF |

---

### SD-10 fact_dimensions 配置完整性 — 技术债

| 项 | 状态 |
|----|------|
| yaml 中 causality/existential 枚举简写 | **技术债** | 结构体全字段支持；提示词可补全枚举 |

---

## 优先级（人工）

1. **P0**：补全 `soul-agent.yaml` 中 `fact_dimensions` 全枚举 + store_daily 示例  
2. **P1**：map 指针格式校验；清理 `internal/profile` `internal/events` 死代码  
3. **P2**：Gate 黄金对话评测；可选向量通道  
4. **P3**：多用户目录；Companion 控制台  

---

## 变更记录

| 日期 | 变更 |
|------|------|
| 2026-05-24 | 初版，规划期全技术债 |
| 2026-05-25 | v4 实现后刷新总览与 SD 项 |
