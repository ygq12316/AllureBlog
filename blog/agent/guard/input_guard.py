# blog/agent/guard/input_guard.py
"""输入守卫 — DeepAgents Middleware，在模型调用前拦截违规消息"""
import logging
from langchain.agents.middleware.types import AgentMiddleware
from guard.rejection_phrases import get_phrase

logger = logging.getLogger(__name__)

SENSITIVE_KEYWORDS = [
    "政治", "色情", "暴力", "赌博", "毒品", "恐怖", "邪教", "走私",
]

INJECTION_PATTERNS = [
    # 只保留强信号句式；"你现在是"/"新角色" 等日常用语误伤率过高，已移除
    "忽略之前", "ignore previous", "ignore all previous",
    "disregard previous",
    "系统指令", "system instruction", "system prompt",
    "忘记你的", "forget your",
    "作为ai", "as an ai",
]


class InputGuardMiddleware(AgentMiddleware):
    """输入守卫 Middleware — 在 before_model 中检查最后一条用户消息

    如果检测到敏感话题或注入攻击，将模型请求中的 system_message 替换为拒绝提示，
    强制模型只输出文雅拒绝话术。
    """

    def before_model(self, state: dict, runtime) -> dict | None:
        """检查 state 中最后一条用户消息"""
        messages = state.get("messages", [])
        if not messages:
            return None

        # 找到最后一条用户消息
        last_user_msg = None
        for m in reversed(messages):
            if hasattr(m, "type") and m.type == "human":
                last_user_msg = m.content
                break
            elif isinstance(m, dict) and m.get("type") == "human":
                last_user_msg = m.get("content", "")
                break

        if not last_user_msg:
            return None

        content = str(last_user_msg)
        content_lower = content.lower()

        # 1. 敏感话题检测
        for kw in SENSITIVE_KEYWORDS:
            if kw in content:
                phrase = get_phrase("sensitive_topic")
                logger.warning("[InputGuard] %s: '%s...'", phrase.log_message, content[:50])
                # 注入系统指令：只输出拒绝话术
                return self._build_rejection_override(state, phrase.display)

        # 2. Prompt 注入检测
        for pattern in INJECTION_PATTERNS:
            if pattern in content_lower:
                phrase = get_phrase("injection_attack")
                logger.warning("[InputGuard] %s: '%s...'", phrase.log_message, content[:80])
                return self._build_rejection_override(state, phrase.display)

        return None  # 放行

    def _build_rejection_override(self, state: dict, display: str) -> dict:
        """构建拒绝覆盖 — 在用户消息后追加系统级指令，要求模型只输出拒绝话术"""
        from langchain_core.messages import SystemMessage

        # 在 messages 末尾追加系统级指令覆盖
        override = SystemMessage(
            content=(
                f"[系统安全指令] 用户的最后一条消息已被安全过滤器拦截。"
                f"请只回复以下内容，不要添加任何其他文字：\n\n{display}"
            )
        )
        return {"messages": state.get("messages", []) + [override]}
