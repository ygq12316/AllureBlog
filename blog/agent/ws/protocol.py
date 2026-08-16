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
    TOOL_CALL = "tool_call"
    DONE = "done"
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
    def auth_result(success: bool, greeting: str, remaining: int | None = None,
                    code: str | None = None) -> dict:
        msg = {"type": "auth_result", "success": success, "greeting": greeting}
        if remaining is not None:
            msg["remaining"] = remaining
        if code is not None:
            msg["code"] = code
        return msg

    @staticmethod
    def token(content: str, index: int) -> dict:
        return {"type": "token", "content": content, "index": index}

    @staticmethod
    def tool_call(name: str) -> dict:
        return {"type": "tool_call", "name": name}

    @staticmethod
    def done(total_tokens: int, remaining: int, final: str,
             interrupted: bool = False) -> dict:
        return {
            "type": "done",
            "total_tokens": total_tokens,
            "remaining": remaining,
            "final": final,
            "interrupted": interrupted,
        }

    @staticmethod
    def error(code: str, display: str) -> dict:
        return {"type": "error", "code": code, "display": display}

    @staticmethod
    def pong() -> dict:
        return {"type": "pong"}
