"""输出过滤 — 纯函数，打码隐私信息、修正角色破坏性开头"""
import re
import logging

logger = logging.getLogger(__name__)

PHONE_PATTERN = re.compile(r"1[3-9]\d{9}")
EMAIL_PATTERN = re.compile(r"[\w.\-]+@[\w.\-]+\.\w+")

BAD_STARTS = ["作为一个ai", "作为人工智能", "作为语言模型", "作为一个语言", "作为一个智能"]


def filter_output(text: str) -> str:
    """过滤最终回答：手机号/邮箱打码；AI 自称开头改写为精灵口吻"""
    if not text:
        return text

    text = PHONE_PATTERN.sub("***", text)
    text = EMAIL_PATTERN.sub("***@***", text)

    text_lower = text.lower()
    for start in BAD_STARTS:
        if text_lower.startswith(start):
            text = "笔墨以为，" + text[len(start):]
            break

    return text.strip()
