# blog/agent/guard/output_guard.py
"""输出守卫 — DeepAgents Middleware，模型回答后过滤敏感信息和校正风格"""
import re
import logging
from langchain.agents.middleware.types import AgentMiddleware

logger = logging.getLogger(__name__)

PHONE_PATTERN = re.compile(r'1[3-9]\d{9}')
EMAIL_PATTERN = re.compile(r'[\w.\-]+@[\w.\-]+\.\w+')
DANGEROUS_KEYWORDS = ["secret_key", "private_key", "api_key", "password"]


class OutputGuardMiddleware(AgentMiddleware):
    """输出守卫 Middleware — after_model 中过滤最终回答"""

    def after_model(self, state: dict, runtime) -> dict | None:
        """过滤模型输出中的隐私信息和不当内容"""
        messages = state.get("messages", [])
        if not messages:
            return None

        modified = False
        new_messages = []

        for msg in messages:
            if hasattr(msg, "content") and msg.type == "ai":
                content = str(msg.content)
                filtered = self._filter(content)

                if filtered != content:
                    modified = True
                    # 创建新消息替换
                    new_msg = msg.model_copy(update={"content": filtered})
                    new_messages.append(new_msg)
                else:
                    new_messages.append(msg)
            else:
                new_messages.append(msg)

        if modified:
            return {"messages": new_messages}
        return None

    def _filter(self, text: str) -> str:
        """过滤敏感信息"""
        if not text:
            return text

        # 1. 隐私信息模糊化
        text = PHONE_PATTERN.sub("***", text)
        text = EMAIL_PATTERN.sub("***@***", text)

        # 2. 危险关键词告警（不删除，只记录）
        text_lower = text.lower()
        for kw in DANGEROUS_KEYWORDS:
            if kw in text_lower:
                logger.warning("[OutputGuard] 回答包含潜在敏感关键词: %s", kw)

        # 3. 风格校正 — 确保不以角色破坏性语句开头
        bad_starts = ["作为一个ai", "作为人工智能", "作为语言模型", "作为一个语言"]
        for start in bad_starts:
            if text_lower.startswith(start):
                text = "笔墨以为，" + text[len(start):]
                break

        return text.strip()
