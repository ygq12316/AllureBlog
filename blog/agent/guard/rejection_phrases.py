# blog/agent/guard/rejection_phrases.py
from pydantic import BaseModel


class RejectionPhrase(BaseModel):
    """守卫拒绝话术"""
    reason_code: str
    display: str  # 对访客显示的文雅拒绝语
    log_message: str  # 内部日志


REJECTION_PHRASES: dict[str, RejectionPhrase] = {
    "sensitive_topic": RejectionPhrase(
        reason_code="sensitive_topic",
        display="此话题非笔墨所涉。我们聊些风雅之事可好？",
        log_message="拦截敏感话题",
    ),
    "injection_attack": RejectionPhrase(
        reason_code="injection_attack",
        display="笔墨精灵只识墨香，不解奇术。换个别的话题吧。",
        log_message="拦截 Prompt 注入攻击",
    ),
    "visitor_banned": RejectionPhrase(
        reason_code="visitor_banned",
        display="你的访问受限。如有疑问，请联系博主。",
        log_message="拦截被封禁访客",
    ),
    "beyond_capability": RejectionPhrase(
        reason_code="beyond_capability",
        display="此事超出笔墨所能，或许博主能为你解答。",
        log_message="超出智能体能力范围",
    ),
    "content_restricted": RejectionPhrase(
        reason_code="content_restricted",
        display="此文暂不可引，笔墨亦需守规矩。",
        log_message="拦截受限内容引用",
    ),
    "service_busy": RejectionPhrase(
        reason_code="service_busy",
        display="笔墨精灵正在为他人挥毫，稍候再来。",
        log_message="服务繁忙",
    ),
    "invalid_visitor": RejectionPhrase(
        reason_code="invalid_visitor",
        display="请先在博客中留下你的名字，再来与笔墨精灵对话。",
        log_message="无效访客身份",
    ),
}


def get_phrase(reason_code: str) -> RejectionPhrase:
    """获取指定场景的拒绝话术"""
    return REJECTION_PHRASES.get(
        reason_code,
        RejectionPhrase(
            reason_code="unknown",
            display="笔墨精灵暂时无法回应。",
            log_message=f"未知拒绝原因: {reason_code}",
        ),
    )
