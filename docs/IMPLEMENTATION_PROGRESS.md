# Soul MCP — 实现进度

> **phase**: `4-async-pipeline`  
> **最后更新**: 2026-05-25

---

## 里程碑

| 里程碑 | 状态 | 备注 |
|--------|------|------|
| M0 设计文档（DESIGN / ARCH / ACCEPTANCE / DRIFT） | ✅ | v4 已同步 |
| M1 `cmd/soul-mcp` + 双工具 | ✅ | stdio |
| M2 按天落地 `storage/history/*.jsonl` | ✅ | 四维 Fact |
| M3 四路异步 store（daily/person/map/prefetch） | ✅ | 独立 LLM 提示 |
| M4 `person.md` + `map.md` + `llm_cache.json` | ✅ | |
| M5 快慢双轨 retrieve（Gate + Compose） | ✅ | 无 LLM 模板降级 |
| M6 `internal/recall` 多通道 + RRF | ✅ | 含「昨天」时间窗 |
| M7 `soul.agent.yaml` 只读 + hints 绑定灵魂 | ✅ | `SOUL_MCP_SOUL_DOC` |
| M8 Host `plan_soul_hook` + 边界测试 | ✅ | AgentTest 仓库 |
| M9 旧 profile/events 迁移工具 | ⬜ | 残留数据可手动清理 |
| M10 Gate 黄金集 / 向量检索 | ⬜ | 见 DRIFT SD-5/9 |

---

## 已 superseded

| 旧里程碑 | 说明 |
|----------|------|
| profile.jsonl / events.jsonl 主路径 | v4 不再写入 |
| `2-llm-triad` 单文件 history + 单次 store LLM | 由 v4 取代 |
| `soul.config` + overlay | 由 soul.agent.yaml + person.md 取代 |
| retrieve `persona_prompt` + `event_context` 双字段 | Host 统一用 `hints` |

---

## 代码入口速查

| 能力 | 文件 |
|------|------|
| Store 编排 | `internal/engine/store_pipeline.go` |
| Retrieve 编排 | `internal/engine/retrieve_pipeline.go` |
| LLM 任务 | `internal/soulagent/store_tasks.go`, `retrieve_pipeline.go` |
| 配置 | `agentConfig/soul-agent.yaml` |

---

## 验收

```bash
go test ./...
```

Host：`AgentTest/scripts/soul_boundary_test`（需 `plan_soul_hook.enabled`）。

---

## 开放 PR / 分支

（无）
