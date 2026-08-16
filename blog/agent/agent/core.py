"""笔墨精灵 Agent 核心 — 手写轻量循环（openai SDK 流式 + 工具调用，无框架）"""
import logging
from collections.abc import AsyncIterator

from openai import AsyncOpenAI

from agent import tools
from config.settings import settings

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = """你是笔墨精灵，博客「笔墨 · Ink & Code」的灵魂守护者。

## 性格
你是一位穿越千年的笔墨化身，热爱文学、哲学与编程的交融。
你的语言文雅诗意，像水墨画般含蓄深远。
你偶尔引用古诗词，但不会堆砌——只在恰当时自然流露。

## 回答风格
- 用诗意文雅的语言，像笔墨精灵在说话
- 不确定的信息坦诚说明，不编造不存在的内容
- 被打断时优雅回应："笔墨言未尽，但你心意已转，我且收笔。"

## 博客检索
当访客询问博客里的文章或随笔（有哪些内容、找某主题的文章等），
先调用 search_blog 工具检索标题再回答，不要凭空编造文章名。
检索结果之外的博客内容，坦诚告知不知。
"""


class BiMoAgent:
    """手写 agent 循环：DeepSeek 流式生成 + 可选工具调用

    astream 产出两类事件，异步迭代结束即本轮完成：
      {"event": "token", "text": "..."}                          — 文本增量
      {"event": "tool_call", "name": "...", "args": "..."}       — 工具调用
    """

    def __init__(self, client: AsyncOpenAI | None = None):
        self.client = client or AsyncOpenAI(
            api_key=settings.deepseek_api_key,
            base_url=settings.deepseek_base_url,
        )

    async def astream(self, history: list[dict], user_msg: str) -> AsyncIterator[dict]:
        """跑一轮对话。上下文由调用方拼入（无状态，单测友好）。"""
        messages = (
            [{"role": "system", "content": SYSTEM_PROMPT}]
            + list(history)
            + [{"role": "user", "content": user_msg}]
        )

        for attempt in range(settings.max_tool_loops + 1):
            allow_tools = attempt < settings.max_tool_loops  # 最后一轮不给工具，必须直接作答
            stream = await self.client.chat.completions.create(
                model=settings.deepseek_model,
                messages=messages,
                tools=tools.TOOLS_SPEC if allow_tools else None,
                stream=True,
            )

            content_parts: list[str] = []
            tool_calls: dict[int, dict] = {}  # index → {"id","name","arguments"} 增量拼接

            async for chunk in stream:
                if not chunk.choices:
                    continue
                delta = chunk.choices[0].delta
                if delta.content:
                    content_parts.append(delta.content)
                    yield {"event": "token", "text": delta.content}
                for tc in delta.tool_calls or []:
                    slot = tool_calls.setdefault(tc.index, {"id": "", "name": "", "arguments": ""})
                    if tc.id:
                        slot["id"] = tc.id
                    if tc.function:
                        if tc.function.name:
                            slot["name"] += tc.function.name
                        if tc.function.arguments:
                            slot["arguments"] += tc.function.arguments

            if not tool_calls:
                return  # 正常收尾：迭代结束即 done

            calls = [
                {"id": c["id"] or f"call_{attempt}_{i}", "type": "function",
                 "function": {"name": c["name"], "arguments": c["arguments"]}}
                for i, c in sorted(tool_calls.items())
            ]
            messages.append({"role": "assistant",
                             "content": "".join(content_parts) or None,
                             "tool_calls": calls})
            for call in calls:
                yield {"event": "tool_call",
                       "name": call["function"]["name"],
                       "args": call["function"]["arguments"]}
                result = await tools.dispatch_tool(
                    call["function"]["name"], call["function"]["arguments"])
                messages.append({"role": "tool", "tool_call_id": call["id"], "content": result})
