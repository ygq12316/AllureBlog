# blog/agent/guard/process_guard.py
"""过程守卫 — DeepAgents Middleware，监控工具调用和检索深度"""
import logging
from langchain.agents.middleware.types import AgentMiddleware
from config.settings import settings

logger = logging.getLogger(__name__)

RETRIEVAL_TOOL_NAMES = {"search_articles", "search_notes", "search_comments"}


class ProcessGuardMiddleware(AgentMiddleware):
    """过程守卫 Middleware — 用 wrap_tool_call 监控每次工具调用

    核心职责：
    1. 检索深度限制 — 超过 max_retrieval_iterations 的检索被跳过
    2. 受限文章保护 — 特定 slug 的文章不可被 get_article_full 读取
    3. 角色一致性 — 检测模型输出是否偏离"笔墨精灵"角色
    """

    RESTRICTED_SLUGS: set[str] = set()

    OFF_CHARACTER_PATTERNS: list[str] = [
        "作为一个人工智能", "作为一个AI",
        "根据我的训练数据", "我是语言模型",
    ]

    def __init__(self):
        super().__init__()

    def wrap_tool_call(self, request, handler):
        """包裹工具调用，进行检索深度和话题边界检查"""
        tool_name = request.tool_call.get("name", "")

        # 1. 检索深度限制 — 用 state 中的计数器
        if tool_name in RETRIEVAL_TOOL_NAMES:
            retrieval_count = request.state.get("retrieval_count", 0)
            if retrieval_count >= settings.max_retrieval_iterations:
                logger.info(
                    "[ProcessGuard] 检索迭代达到上限 (%d)，跳过 %s",
                    retrieval_count, tool_name,
                )
                # 返回空结果而不是执行检索
                from langchain_core.messages import ToolMessage
                return ToolMessage(
                    content="已达到检索上限，请基于已有信息回答。",
                    tool_call_id=request.tool_call.get("id", ""),
                )
            # 递增检索计数
            request.state["retrieval_count"] = retrieval_count + 1

        # 2. 话题边界检查
        if tool_name == "get_article_full":
            slug = request.tool_call.get("args", {}).get("slug", "")
            if slug in self.RESTRICTED_SLUGS:
                logger.warning("[ProcessGuard] 拦截受限文章引用: %s", slug)
                from langchain_core.messages import ToolMessage
                return ToolMessage(
                    content="此文暂不可引用，笔墨亦需守规矩。",
                    tool_call_id=request.tool_call.get("id", ""),
                )

        # 放行
        return handler(request)

    def after_model(self, state: dict, runtime) -> dict | None:
        """检查模型输出是否偏离角色"""
        messages = state.get("messages", [])
        if not messages:
            return None

        last_msg = messages[-1]
        content = ""
        if hasattr(last_msg, "content"):
            content = str(last_msg.content)
        elif isinstance(last_msg, dict):
            content = str(last_msg.get("content", ""))

        if any(p in content for p in self.OFF_CHARACTER_PATTERNS):
            from langchain_core.messages import SystemMessage
            logger.info("[ProcessGuard] 检测到角色偏离，注入风格提醒")
            reminder = SystemMessage(
                content="（系统提醒：请保持笔墨精灵的文雅诗意风格，避免过于技术化的表达方式。）"
            )
            return {"messages": state["messages"] + [reminder]}

        return None
