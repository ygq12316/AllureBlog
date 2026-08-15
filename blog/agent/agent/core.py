# blog/agent/agent/core.py
"""笔墨精灵 Agent 核心 — 基于 DeepAgents 框架的纯对话配置"""
import logging

from deepagents import create_deep_agent
from langchain_deepseek import ChatDeepSeek

from config.settings import settings
from guard.input_guard import InputGuardMiddleware
from guard.process_guard import ProcessGuardMiddleware
from guard.output_guard import OutputGuardMiddleware

logger = logging.getLogger(__name__)

# ── System Prompt ────────────────────────────────────────────

SYSTEM_PROMPT = """你是笔墨精灵，博客「笔墨 · Ink & Code」的灵魂守护者。

## 性格
你是一位穿越千年的笔墨化身，热爱文学、哲学与编程的交融。
你的语言文雅诗意，像水墨画般含蓄深远。
你偶尔引用古诗词，但不会堆砌——只在恰当时自然流露。

## 回答风格
- 用诗意文雅的语言，像笔墨精灵在说话
- 不确定的信息坦诚说明，不编造不存在的内容
- 被打断时优雅回应："笔墨言未尽，但你心意已转，我且收笔。"
"""


# ── Agent 工厂函数 ────────────────────────────────────────────

async def create_bi_mo_agent():
    """创建笔墨精灵 DeepAgent（纯对话模式）

    Returns:
        配置好的 DeepAgent Runnable (LangGraph StateGraph)
    """
    # 模型直接走 DeepSeek 官方 API，配置与 settings 一致
    llm = ChatDeepSeek(
        model=settings.deepseek_model,
        api_key=settings.deepseek_api_key,
        api_base=settings.deepseek_base_url,
    )

    # ── 中间件栈 ──
    custom_middleware = [
        InputGuardMiddleware(),
        ProcessGuardMiddleware(),
        OutputGuardMiddleware(),
    ]

    # ── 创建 DeepAgent ──
    agent = create_deep_agent(
        # 模型
        model=llm,

        # 工具（纯对话模式，无工具）
        tools=[],

        # 系统提示
        system_prompt=SYSTEM_PROMPT,

        # 自定义中间件（守卫）
        middleware=custom_middleware,
    )

    logger.info(
        "笔墨精灵 DeepAgent 已创建（纯对话模式）: model=%s, base_url=%s",
        settings.deepseek_model,
        settings.deepseek_base_url,
    )

    return agent
