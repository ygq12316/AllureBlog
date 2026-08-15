# blog/agent/ws/protocol.py
from enum import StrEnum
from pydantic import BaseModel


class ClientMessageType(StrEnum):
    """客户端 → 服务端 消息类型"""
    AUTH = "auth"
    MESSAGE = "message"
    INTERRUPT = "interrupt"
    PING = "ping"


class ServerMessageType(StrEnum):
    """服务端 → 客户端 消息类型"""
    AUTH_RESULT = "auth_result"
    TOKEN = "token"
    DONE = "done"
    REJECTED = "rejected"
    ERROR = "error"
    PONG = "pong"


class ClientMessage(BaseModel):
    """客户端发来的消息"""
    type: str
    visitor_uuid: str | None = None
    visitor_name: str | None = None
    content: str | None = None


class ServerMessage:
    """服务端消息工厂"""

    @staticmethod
    def auth_result(success: bool, greeting: str) -> dict:
        return {"type": "auth_result", "success": success, "greeting": greeting}

    @staticmethod
    def token(content: str, index: int) -> dict:
        return {"type": "token", "content": content, "index": index}

    @staticmethod
    def done(total_tokens: int, sources: list[dict], interrupted: bool = False) -> dict:
        return {
            "type": "done",
            "total_tokens": total_tokens,
            "sources": sources,
            "interrupted": interrupted,
        }

    @staticmethod
    def rejected(reason: str, display: str) -> dict:
        return {"type": "rejected", "reason": reason, "display": display}

    @staticmethod
    def error(code: str, display: str) -> dict:
        return {"type": "error", "code": code, "display": display}

    @staticmethod
    def pong() -> dict:
        return {"type": "pong"}