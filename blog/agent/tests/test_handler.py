import fakeredis.aioredis
import httpx
import pytest
from fastapi import FastAPI, WebSocket
from fastapi.testclient import TestClient
from starlette.websockets import WebSocketDisconnect

from memory.chat_store import ChatStore
from ws.handler import ChatHandler

ACCOUNT_VISITOR = {"uuid": "acct1", "username": "lin", "nickname": "林间客"}
ANON_VISITOR = {"uuid": "anon1", "username": "", "nickname": "访客1234"}


class FakeAgent:
    """打桩 BiMoAgent：每轮吐固定事件序列，记录收到的 history"""

    def __init__(self, events):
        self.events = events
        self.seen_history = None

    async def astream(self, history, user_msg):
        self.seen_history = history
        for e in self.events:
            yield e


def make_client(fake_agent, visitor, monkeypatch):
    app = FastAPI()
    store = ChatStore(redis=fakeredis.aioredis.FakeRedis(decode_responses=True))
    handler = ChatHandler(agent=fake_agent, store=store)

    async def fake_fetch(self, uuid, client=None):
        return visitor

    # monkeypatch 打桩登录校验，用例结束自动还原（不污染其它用例的类属性）
    monkeypatch.setattr(ChatHandler, "_fetch_visitor", fake_fetch)

    @app.websocket("/chat/ws")
    async def ws_endpoint(ws: WebSocket):
        await handler.handle(ws)

    return TestClient(app), store


def recv_until(ws, msg_type, limit=20):
    """收消息直到指定类型，返回该消息"""
    for _ in range(limit):
        data = ws.receive_json()
        if data["type"] == msg_type:
            return data
    raise AssertionError(f"未收到 {msg_type}")


def test_anonymous_visitor_rejected(monkeypatch):
    agent = FakeAgent([])
    client, _ = make_client(agent, ANON_VISITOR, monkeypatch)
    with client.websocket_connect("/chat/ws") as ws:
        ws.send_json({"type": "auth", "visitor_uuid": "anon1", "visitor_name": "x"})
        result = ws.receive_json()
        assert result["type"] == "auth_result"
        assert result["success"] is False
        assert result["code"] == "login_required"
        with pytest.raises(WebSocketDisconnect):
            ws.receive_json()  # 服务端随后关闭连接


def test_account_visitor_chat_round_persists_history(monkeypatch):
    agent = FakeAgent([
        {"event": "token", "text": "幸会"},
        {"event": "token", "text": "，久仰"},
    ])
    client, store = make_client(agent, ACCOUNT_VISITOR, monkeypatch)
    with client.websocket_connect("/chat/ws") as ws:
        ws.send_json({"type": "auth", "visitor_uuid": "acct1", "visitor_name": "lin"})
        result = ws.receive_json()
        assert result["success"] is True
        assert result["remaining"] == 10
        assert "林间客" in result["greeting"]  # 问候用 Go 记录里的昵称

        ws.send_json({"type": "message", "content": "你好"})
        t1 = ws.receive_json()
        assert t1 == {"type": "token", "content": "幸会", "index": 0}
        done = recv_until(ws, "done")
        assert done["final"] == "幸会，久仰"
        assert done["remaining"] == 9

    # 本轮已入历史；第二轮 agent 能拿到上一轮（多轮记忆生效）
    assert agent.seen_history == []
    with client.websocket_connect("/chat/ws") as ws:
        ws.send_json({"type": "auth", "visitor_uuid": "acct1"})
        ws.receive_json()
        ws.send_json({"type": "message", "content": "再来一句"})
        recv_until(ws, "done")
    assert agent.seen_history == [
        {"role": "user", "content": "你好"},
        {"role": "assistant", "content": "幸会，久仰"},
    ]


def test_daily_limit_exceeded(monkeypatch):
    agent = FakeAgent([{"event": "token", "text": "答"}])
    client, _ = make_client(agent, ACCOUNT_VISITOR, monkeypatch)
    with client.websocket_connect("/chat/ws") as ws:
        ws.send_json({"type": "auth", "visitor_uuid": "acct1"})
        ws.receive_json()
        # 前 10 条正常应答（计数从 0 起，auth 后 remaining=10）
        for i in range(10):
            ws.send_json({"type": "message", "content": f"第{i}问"})
            done = recv_until(ws, "done")
            assert done["remaining"] == 9 - i
        # 第 11 条被拒
        ws.send_json({"type": "message", "content": "第11问"})
        err = recv_until(ws, "error")
        assert err["code"] == "limit_exceeded"
        assert "笔墨已尽" in err["display"]


def test_sensitive_input_short_circuits(monkeypatch):
    agent = FakeAgent([])  # 不应被调用
    client, _ = make_client(agent, ACCOUNT_VISITOR, monkeypatch)
    with client.websocket_connect("/chat/ws") as ws:
        ws.send_json({"type": "auth", "visitor_uuid": "acct1"})
        ws.receive_json()
        ws.send_json({"type": "message", "content": "聊聊赌博"})
        done = recv_until(ws, "done")
        assert "不宜深谈" in done["final"]
    assert agent.seen_history is None  # agent 未被触达


def test_tool_call_event_forwarded(monkeypatch):
    agent = FakeAgent([
        {"event": "tool_call", "name": "search_blog", "args": '{"keyword": ""}'},
        {"event": "token", "text": "博客里有…"},
    ])
    client, _ = make_client(agent, ACCOUNT_VISITOR, monkeypatch)
    with client.websocket_connect("/chat/ws") as ws:
        ws.send_json({"type": "auth", "visitor_uuid": "acct1"})
        ws.receive_json()
        ws.send_json({"type": "message", "content": "博客有什么文章"})
        tc = recv_until(ws, "tool_call")
        assert tc["name"] == "search_blog"
        recv_until(ws, "done")


async def test_fetch_visitor_status_semantics():
    """登录校验的返回值语义：200→访客 dict；404→{}（未登录）；5xx/网络错→None（fail-closed）"""
    handler = ChatHandler(agent=FakeAgent([]), store=None)

    def transport(request: httpx.Request) -> httpx.Response:
        tail = request.url.path.rsplit("/", 1)[-1]
        if tail == "ok":
            return httpx.Response(200, json={"visitor": ACCOUNT_VISITOR})
        if tail == "anon":
            return httpx.Response(200, json={"visitor": ANON_VISITOR})
        if tail == "missing":
            return httpx.Response(404, json={"error": "未找到"})
        return httpx.Response(500, json={"error": "boom"})

    client = httpx.AsyncClient(transport=httpx.MockTransport(transport), timeout=3)
    assert await handler._fetch_visitor("ok", client=client) == ACCOUNT_VISITOR
    assert await handler._fetch_visitor("anon", client=client) == ANON_VISITOR
    assert await handler._fetch_visitor("missing", client=client) == {}   # 404 → 未登录
    assert await handler._fetch_visitor("boom", client=client) is None   # 5xx → fail-closed
    await client.aclose()
