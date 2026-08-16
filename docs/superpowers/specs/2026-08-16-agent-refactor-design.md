# 笔墨精灵 Agent 重构设计

日期：2026-08-16
分支：feature/fairy-chat-frontend
状态：已与用户逐节确认

## 1. 背景与动机

现有 `blog/agent/` 基于 deepagents 框架实现纯对话精灵，存在三个实质问题：

1. **多轮对话实际不工作** — `create_deep_agent` 未配置 checkpointer，handler 传入的
   `thread_id` 形同虚设；每次请求只携带单条用户消息，精灵不记得上一轮内容。Redis
   中存了对话历史但从不读回。
2. **deepagents 框架过重** — 它面向多步规划/子代理场景，本项目只用纯对话，框架
   大部分能力闲置；且本机 docker build 一直未跑通。
3. **三层守卫偏复杂** — 输入/过程/输出三套 `AgentMiddleware` 职责部分重叠，角色
   漂移检测、受限文章保护等实际用不上。

同时用户提出新需求：多轮对话、检索博客文章/随笔标题的工具、账号登录门槛、每日
对话次数限制。

## 2. 需求与决策记录

用户原始需求：

1. 重构现有 agent 后端逻辑
2. 多轮对话助手；工具为检索本站文章与随笔标题（访客想查询时）；仅登录后可使
   用；每日对话次数限制 10 次
3. 完善设计为简单助手，可搜索资料
4. 切合当前项目，不过于复杂

逐项澄清的决策（用户均确认）：

| 议题 | 决策 |
|---|---|
| 「登录」的含义 | 访客账号登录（`/api/visitor/register` + `/api/visitor/login` 体系），非匿名访客、非管理员 |
| 次数限制周期 | 每自然日 10 条用户消息，次日重置 |
| 「搜索资料」的范围 | 仅检索博客内文章/随笔标题，不联网 |
| 实现方案 | A：手写轻量 agent 循环，移除 deepagents 与 langchain 全家桶 |

选方案 A 的理由：现有核心 bug（thread_id 无效）暴露了框架黑盒的代价；「多轮对话
+ 1 个检索工具」的助手手写循环一眼看穿，依赖最少，docker build 遗留问题随
deepagents 依赖链一起消失；切合本项目 Go 后端轻依赖手工组装的气质。

## 3. 总体架构

```
浏览器 useFairyChat ──WebSocket──▶ FastAPI (blog/agent)
                                    ├─ ws/handler.py    连接生命周期、登录校验、次数限制
                                    ├─ agent/core.py    手写 agent 循环（openai SDK 流式）
                                    │    └─ tools.py    search_blog：httpx → Go API 检索标题
                                    ├─ memory/chat_store.py  Redis：多轮历史 + 每日计数
                                    └─ guard/ 两个纯函数：输入拦截、输出过滤
                                                │
                              Go 后端 :8080 ◀────┘ GET /api/visitor/:uuid（登录校验）
                                                  GET /api/articles、/api/notes（标题检索）
```

一次对话的完整数据流：

1. 前端 connect → 检查已登录（无 `account` 则弹登录框，不建立连接）
2. auth（带 visitor_uuid）→ agent 调 Go `/api/visitor/:uuid`：访客不存在或
   `username` 为空 → 拒绝；通过 → `auth_result(success=true, greeting, remaining=N)`
3. 前端发 message → `INCR` 当日计数，超 10 → `error(limit_exceeded)`；通过 →
   进入 agent 循环
4. 循环：Redis 取最近 5 轮历史 + system prompt + 本条 → DeepSeek 流式生成 →
   需要检索则调 `search_blog` 再生成（最多 2 轮工具循环）→ token 流式推送 →
   `done(remaining=N, final=过滤后全文)` → 整轮写回 Redis

## 4. 目录与依赖变化

| 现有 | 重构后 |
|---|---|
| `agent/core.py`（deepagents 工厂） | `agent/core.py`（BiMoAgent 手写循环）+ `agent/tools.py`（新增） |
| `guard/` 三中间件 + `rejection_phrases.py` | `guard/input_filter.py` + `guard/output_filter.py`（纯函数） |
| `memory/short_term.py`（load/context/persist_summary 从未用上） | `memory/chat_store.py`（历史 + 计数，职责单一） |
| `ws/handler.py`、`ws/protocol.py` | 保留骨架；auth 处加校验，协议加 `remaining`/`final`/`tool_call` |
| 依赖 deepagents、langchain-deepseek、python-dotenv | **移除**；新增 `openai`、`httpx` |

`requirements.txt` 最终：`fastapi`、`uvicorn[standard]`、`redis`、`pydantic`、
`pydantic-settings`、`openai`、`httpx`，测试用 `fakeredis`、`pytest`（注释标明仅测试用）。

## 5. Agent 核心（agent/core.py）

`BiMoAgent` 持有 `openai.AsyncOpenAI` 客户端（指向 DeepSeek 兼容端点），唯一公开
方法为异步生成器：

```python
async def astream(self, history: list[dict], user_msg: str) -> AsyncIterator[dict]:
    # 产出: {"event": "token", "text": "..."}
    #       {"event": "tool_call", "name": "search_blog", "args": {...}}
    #       {"event": "done", "messages": [...]}   # 本轮完整消息（含工具调用记录）
```

循环逻辑：

```python
messages = [{"role": "system", "content": SYSTEM_PROMPT}, *history,
            {"role": "user", "content": user_msg}]

for attempt in range(max_tool_loops + 1):   # 默认 2 次工具机会 + 1 次收尾
    allow_tools = attempt < max_tool_loops   # 最后一轮不传 tools，模型必须直接作答
    stream = await self.client.chat.completions.create(
        model=settings.deepseek_model, messages=messages,
        tools=TOOLS_SPEC if allow_tools else None, stream=True)

    tool_calls = []   # 流式增量拼接：按 index 聚合 function.name / arguments
    async for chunk in stream:
        ...  # delta.content → yield token；delta.tool_calls → 拼进 tool_calls

    if not tool_calls:
        return  # 正常收尾
    messages.append({"role": "assistant", "tool_calls": [...]})
    for tc in tool_calls:
        yield {"event": "tool_call", ...}
        result = await dispatch_tool(tc)      # 目前只有 search_blog
        messages.append({"role": "tool", "tool_call_id": ..., "content": result})
```

设计要点：

- 带工具调用的生成轮若吐了文字，照常流给前端（DeepSeek 调工具时通常不吐字；
  即使吐了，消息记录保留 content，下轮不会重放）。
- 工具上限的收法：最后一轮干脆不传 `tools`，模型只能文字作答，无需特殊错误处理。
- 无状态、无 checkpointer：上下文由调用方（handler）从 Redis 拼入，`astream`
  近似纯函数，单测友好。

## 6. 检索工具（agent/tools.py）

```python
TOOLS_SPEC = [{
    "type": "function",
    "function": {
        "name": "search_blog",
        "description": "检索博客内的文章与随笔标题。访客想了解博客写了什么、找某主题的文章/随笔时调用。",
        "parameters": {"type": "object",
                       "properties": {"keyword": {"type": "string", "description": "标题过滤关键词；留空返回全部"}},
                       "required": []}}}]
```

执行：`httpx` 并行请求 `GET {blog_api_base}/api/articles?per_page=50` 与
`GET {blog_api_base}/api/notes?per_page=50`（Go 侧参数名为 `per_page`；notes 默认
只返回已发布），按关键词大小写不敏感过滤标题，格式化为文本返回，上限 20 条：

```
文章《Vue3 组合式 API 实战》 分类:前端
随笔《深夜写代码的仪式感》
（共 2 条）
```

Go 侧 API 超时 3 秒；失败返回 `"检索暂时不可用"` 文本，模型据此向访客致歉，
不中断循环。

## 7. 登录校验、次数限制、多轮记忆

### 登录校验（handler 的 auth 步骤）

```python
r = await httpx.get(f"{blog_api_base}/api/visitor/{uuid}", timeout=3)
# username 非空 → 账号用户，放行 auth_result(success=true, greeting, remaining=N)
# username 为空 → 匿名访客，auth_result(success=false, code="login_required",
#                  display="请先注册账号并登录，再来与笔墨精灵对话。") 后关闭
# Go API 不可达 → 拒绝（fail-closed，code="auth_failed"，display="暂时无法验证身份"）
```

fail-closed 理由：Go 挂了就放行等于 curl 直连 agent 即可绕过登录校验；个人博客
宁可短暂不可用，不留绕过口子。

身份关系（已核实）：前端登录成功后 `visitor.uuid` 会被切换为账号访客的 uuid，
因此 auth 消息继续携带 `visitor.uuid`，服务端以该记录的 `username` 判定。

### 次数限制（Redis 计数）

- 键：`chat:{uuid}:count:{YYYYMMDD}`，日期取东八区（受众中文访客，服务器跑 UTC
  也不错位）。
- `INCR` 后首增设 `EXPIRE 48h`，隔天键自然过期，无需清理任务。
- 计数时机：message 进入 agent 前 `INCR`；结果 > 10 → `error(limit_exceeded,
  "今日笔墨已尽（10 次），明日再会。")`。被拒的这条也计入额度——原子、防反复试探。
- `remaining = max(0, 10 - 当前计数)`，在 `auth_result` 与每次 `done` 带回。

### 多轮记忆（memory/chat_store.py）

- 键：`chat:{uuid}:history`，LIST，元素为 `{"role","content"}` JSON。
- 每轮 done 后 RPUSH user + assistant 两条 → `LTRIM` 保留最近 10 条（5 轮）→
  `EXPIRE` 30 分钟（半小时无对话即遗忘，沿用现有 TTL 语义）。
- 下一轮循环 `LRANGE` 读出直接拼进 messages。历史只存文本对话；工具调用轮
  （tool_calls / tool 结果）不进历史。
- 被打断的回答：已生成部分照常写入（精灵记得自己说到一半的话）。

```python
class ChatStore:
    async def get_history(uuid) -> list[dict]
    async def append_round(uuid, user: str, assistant: str)
    async def incr_today(uuid) -> int
    async def remaining_today(uuid) -> int
```

## 8. 守卫简化

- `guard/input_filter.py`：`check_input(text) -> str | None`，命中敏感词或注入
  句式（关键词表沿用现有）则返回拒绝话术。handler 顺序：**INCR → 超限拒绝 →
  输入拦截**；被拦截的直接回固定话术，不进模型（计数哲学与 limit 一致：发出的
  message 皆计数）。
- `guard/output_filter.py`：`filter_output(text) -> str`，手机号/邮箱正则打码 +
  角色破坏开头改写（「作为AI…」→「笔墨以为，…」）。作用于完整回答，通过 done
  的 `final` 字段生效——流式 token 已实时上屏，最终以 `final` 定稿覆盖。
- 删除：`process_guard.py`（检索深度由 `max_tool_loops` 天然约束；受限文章/
  角色漂移检测 YAGNI）、`rejection_phrases.py`（并入 input_filter）、配置
  `max_retrieval_iterations`。

## 9. WebSocket 协议（ws/protocol.py）

| 消息 | 变化 |
|---|---|
| 客户端 auth / message / interrupt / ping | 不变 |
| `auth_result` | + `code`（`login_required`/`auth_failed`）+ `remaining` |
| `token` | 不变 |
| `tool_call`（新增） | `{"type":"tool_call","name":"search_blog"}`，前端显示「正在检索博客…」 |
| `done` | + `remaining` + `final`（过滤后定稿文本） |
| `error` | code 固定枚举：`not_ready`/`auth_failed`/`login_required`/`limit_exceeded`/`agent_error` |
| ~~`rejected`~~ | 删除（从未实际使用） |

## 10. 前端改动（useFairyChat.js + FairyChat.vue）

- `connect()` 前查 `account`（useVisitor 已导出）：未登录 → `status='need-login'`，
  聊天窗显示「登录后可与笔墨精灵对话」+ 按钮打开现有登录弹窗（`openLogin()`），
  登录成功后再自动 connect。
- `handle()` 修复：`auth_result.success=false` 时显示 `display` 文案并置 offline
  （现状前端未检查 success）。
- 新分支：`tool_call` → 显示检索状态条；`done` → 用 `m.final` 定稿气泡 + 更新余量。
- 输入框旁常驻「今日余 N 次」；登出时 `disconnect()`。

## 11. 配置与部署

- `config/settings.py` 新增：`blog_api_base="http://localhost:8080"`、
  `daily_message_limit=10`、`max_tool_loops=2`；删除 `max_retrieval_iterations`；
  保留 `deepseek_*`、`redis_url`、`log_level`、`short_memory_ttl_seconds`（30 分钟
  历史 TTL）、`max_history_rounds`（5 轮记忆窗口）。
- `docker-compose.yml`：agent 服务加 `BLOG_API_BASE=http://app:8080`；Caddyfile 不动。
- agent Dockerfile 结构不变，依赖链大幅缩短，deepagents 相关 build 问题随之消失。

## 12. 测试策略

- 单测（pytest）：
  - input_filter 命中/放行；output_filter 打码/改写
  - search_blog 标题过滤（mock httpx 响应）
  - agent 循环：手写 fake OpenAI 流对象，验证 tool_call 增量拼接、工具循环
    上限、final 输出
  - ChatStore（fakeredis）：LTRIM 窗口、INCR 计数、remaining 计算
- 冒烟（smoke.mjs，仅本机）：精灵段改为先 `/api/visitor/register` 注册账号再
  auth，断言 `success=true` 且收完一轮 `done`。
- 手动验收清单：未登录弹引导 → 登录后可聊 → 问「博客有哪些文章」触发检索 →
  连发 10 条见「笔墨已尽」→ 次日恢复（改 Redis 键日期当场可验）。

## 13. 明确不做（YAGNI）

- 联网搜索工具
- checkpointer / 摘要式长期记忆
- 过程守卫（检索深度中间件、受限文章、角色漂移检测）
- `rejected` 协议消息
- 流式输出过程中的实时内容过滤（以 done.final 定稿覆盖代替）
