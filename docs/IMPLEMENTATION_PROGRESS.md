# Soul MCP — 实现进度

> **phase**: `0-stub`  
> **最后更新**: 2026-05-24

---

## 里程碑

| 里程碑 | 状态 | 备注 |
|--------|------|------|
| M0 设计文档（DESIGN / ARCH / ACCEPTANCE / DRIFT） | ✅ 完成 | 本仓库 |
| M1 `cmd/soul-mcp` + MCP 注册双工具 | ✅ 完成 | stdio；`soul_store` / `soul_retrieve` |
| M2 `soul_store` 同步 ACK + 异步队列 + episode 归档 | 🔶 部分 | stub 写 `data/dialogues.jsonl`；无 profile/events Agent |
| M3 rules 抽取 profile/events | ⬜ 未开始 | 无 LLM 可测 |
| M4 `soul_retrieve` 模板组装 + `soul.config` 加载 | 🔶 部分 | stub hints + 只读 `persona.role` |
| M5 可选 store LLM 整理 | ⬜ 未开始 | env 开关 |
| M6 Host `plan/soulhook` + `app.yaml` | ✅ 完成 | `plan_soul_hook`；默认 `enabled: false` |
| M7 双 MCP E2E（Soul + Memory 并行） | ⬜ 未开始 | |
| M8 Companion 控制台（可选） | ⬜ 未开始 | 非 MVP |

---

## 与 Memory MCP 对齐参考

实现时可参考 `AgentTestMemoryMCP`：

- stdio MCP 入口模式
- store 异步 job + 同步 ACK
- retrieve 预算与 `degraded` 元数据
- **不要**复制 factworld / graph / simple 路由

---

## 开放 PR / 分支

（无）

---

## 下一步（建议）

1. Host 与 Soul 联签 `soul_store` / `soul_retrieve` JSON 字段（冻结 §4）。  
2. 初始化 `go mod` + `cmd/soul-mcp` 空工具回显。  
3. 实现 M4 模板 retrieve（可先无 store）。  
4. 实现 M2–M3 store 管道。  
5. 主项目 `plan/soulhook` 与 portal 注入顺序。
