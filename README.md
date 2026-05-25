# AgentTestSoulMCP

外置 **Soul MCP**（v4）：异步四路存入 + 快慢双轨取出。MCP 工具 **`soul_store` / `soul_retrieve` 接口不变**。

## 运行时目录（`SOUL_MCP_DATA_DIR`）

| 路径 | 维护方 | 作用 |
|------|--------|------|
| `storage/history/YYYY-MM-DD.jsonl` | LLM 任务1（按天） | 冷存储：带四维标签的事实 |
| `person.md` | LLM 任务2 | 用户画像（习惯、话术、思维、情绪等） |
| `map.md` | LLM 任务3 | 热索引：摘要 + 指向落地文件 |
| `llm_cache.json` | LLM 任务4 + Go 检索 | 预测问题的预取缓存 |
| `soul.agent.yaml` | **用户只读** | Agent 人格（`SOUL_MCP_SOUL_DOC`） |
| `agentConfig/soul-agent.yaml` | 运维 | 6 段独立 LLM 提示 + 参数 |

旧版 `history.facts.jsonl` 若仍存在，retrieve 时会合并检索（只读兼容）。

## 存入（异步，4 次独立 LLM）

1. **按天落地** → `storage/history/{today}.jsonl`
2. **用户画像** → `person.md`
3. **地图** → `map.md`（含标签与文件指针）
4. **预取**（wave2）：预测 N 个问题（默认 5，可配置）→ 查 map + 落地文件 → `llm_cache.json`

## 取出（快慢双轨）

1. 载入：`llm_cache` + `map.md` + `person.md` + 用户输入 → **Gate LLM**
2. **快通道**：信息足够 → 直接总结（贴合画像）
3. **慢通道**：输出 `retrieval_tags` → Go `recall.Select` 捞落地文件 → **Compose LLM** 总结
4. 最终 hints = **Agent 灵魂（只读）** + Soul 协作提示

## 配置

```yaml
store:
  max_predicted_questions: 5
  map_recent_days: 7
```

环境变量：`SOUL_MCP_LLM_API_BASE`、`SOUL_MCP_LLM_API_KEY`、`SOUL_MCP_DATA_DIR`、`SOUL_MCP_SOUL_DOC`

## 构建

```powershell
go build -o soul-mcp.exe ./cmd/soul-mcp
```

`phase` 字段：`4-async-pipeline`
