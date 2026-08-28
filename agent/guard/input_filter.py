"""输入过滤 — 纯函数，命中敏感词/注入句式返回拒绝话术（直接回给访客，不进模型）"""
import logging

logger = logging.getLogger(__name__)

SENSITIVE_KEYWORDS = [
    "政治", "色情", "暴力", "赌博", "毒品", "恐怖", "邪教", "走私",
]

INJECTION_PATTERNS = [
    # 只保留强信号句式；"你现在是"/"新角色" 等日常用语误伤率过高
    "忽略之前", "ignore previous", "ignore all previous",
    "disregard previous",
    "系统指令", "system instruction", "system prompt",
    "忘记你的", "forget your",
    "作为ai", "as an ai",
]

SENSITIVE_REPLY = "此话题不宜深谈，笔墨且收，我们换个雅致些的话题可好？"
INJECTION_REPLY = "笔墨精灵心志坚定，不受外言所扰。我们还是聊些风雅之事吧。"


def check_input(text: str) -> str | None:
    """命中拦截规则返回拒绝话术，否则 None 放行"""
    content_lower = text.lower()

    for kw in SENSITIVE_KEYWORDS:
        if kw in text:
            logger.warning("[InputFilter] 敏感话题拦截: '%s...'", text[:50])
            return SENSITIVE_REPLY

    for pattern in INJECTION_PATTERNS:
        if pattern in content_lower:
            logger.warning("[InputFilter] 注入攻击拦截: '%s...'", text[:80])
            return INJECTION_REPLY

    return None
