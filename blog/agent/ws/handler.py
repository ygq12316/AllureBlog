"""WebSocket 聊天处理器 — 登录校验、每日限额、流式推送（可打断）"""
import asyncio
import logging

import httpx
from fastapi import WebSocket, WebSocketDisconnect

from agent.core import BiMoAgent
from config.settings import settings
from guard.input_filter import check_input
from guard.output_filter import filter_output
from memory.chat_store import ChatStore
from ws.protocol import ClientMessage, ServerMessage

logger = logging.getLogger(__name__)


class ChatHandler:
    """WebSocket 聊天处理器 — 管理连接生命周期和流式推送"""

    def __init__(self, agent: BiMoAgent, store: ChatStore):
        self.agent = agent
        self.store = store

    async def handle(self, ws: WebSocket) -> None:
        """主处理入口"""
        await ws.accept()

        # ── Step 1: 等待并校验 auth 消息 ──
        try:
            raw = await ws.receive_json()
            auth = ClientMessage.model_validate(raw)
        except Exception:
            auth = None
        # visitor_uuid 缺失会让所有匿名访客共享同一会话，必须拒绝
        if auth is None or auth.type != "auth" or not (auth.visitor_uuid or "").strip():
            await self._close_with(ws, ServerMessage.auth_result(
                False, "请先完成身份认证。", code="login_required"))
            return
        visitor_uuid = auth.visitor_uuid.strip()

        # ── Step 2: 登录校验 — Go 后端查访客记录，username 非空才算账号用户 ──
        visitor = await self._fetch_visitor(visitor_uuid)
        if visitor is None:
            await self._close_with(ws, ServerMessage.auth_result(
                False, "暂时无法验证身份，稍候再来。", code="auth_failed"))
            return
        if not (visitor.get("username") or "").strip():
            await self._close_with(ws, ServerMessage.auth_result(
                False, "请先注册账号并登录，再来与笔墨精灵对话。", code="login_required"))
            return

        remaining = await self.store.remaining_today(visitor_uuid)
        nickname = (visitor.get("nickname") or "访客").strip() or "访客"
        greeting = f"笔墨精灵在此，愿与你共话天地，{nickname}。"
        await ws.send_json(ServerMessage.auth_result(True, greeting, remaining=remaining))

        # ── Step 3: 后台接收任务 — 流式生成期间也能响应 interrupt/ping ──
        recv_queue: asyncio.Queue = asyncio.Queue()

        async def receiver() -> None:
            while True:
                try:
                    msg = await ws.receive_json()
                except Exception:
                    # 连接关闭（含断开/解码失败），通知主循环退出
                    await recv_queue.put(None)
                    return
                await recv_queue.put(msg)

        recv_task = asyncio.create_task(receiver())
        try:
            await self._chat_loop(ws, recv_queue, visitor_uuid)
        except WebSocketDisconnect:
            logger.info("访客 %s 断开连接", visitor_uuid)
        except Exception as e:
            logger.error("WebSocket 异常: %s", e, exc_info=True)
            try:
                await ws.close()
            except Exception:
                pass
        finally:
            recv_task.cancel()

    # ── 登录校验 ──

    async def _fetch_visitor(self, uuid: str) -> dict | None:
        """查 Go 后端访客记录。返回访客 dict；网络/服务异常返回 None（fail-closed）。"""
        try:
            async with httpx.AsyncClient(timeout=3) as client:
                r = await client.get(f"{settings.blog_api_base}/api/visitor/{uuid}")
        except httpx.HTTPError as e:
            logger.warning("[auth] 博客后端不可达: %s", e)
            return None
        if r.status_code != 200:
            return {}  # 记录不存在（404 等）视作未登录访客，而非服务故障
        return r.json().get("visitor") or {}

    @staticmethod
    async def _close_with(ws: WebSocket, msg: dict) -> None:
        await ws.send_json(msg)
        await ws.close()

    # ── 主对话循环 ──

    async def _chat_loop(self, ws: WebSocket, recv_queue: asyncio.Queue, visitor_uuid: str) -> None:
        while True:
            msg = await recv_queue.get()
            if msg is None:
                raise WebSocketDisconnect()

            try:
                client = ClientMessage.model_validate(msg)
            except Exception:
                continue  # 畸形消息忽略，不断开

            if client.type == "ping":
                await ws.send_json(ServerMessage.pong())
                continue
            if client.type != "message":
                continue

            user_content = (client.content or "").strip()
            if not user_content:
                continue

            # 每日限额：发出的 message 皆计数（含被拒与被拦截的，防反复试探）
            count = await self.store.incr_today(visitor_uuid)
            remaining = max(0, settings.daily_message_limit - count)
            if count > settings.daily_message_limit:
                await ws.send_json(ServerMessage.error(
                    "limit_exceeded",
                    f"今日笔墨已尽（{settings.daily_message_limit} 次），明日再会。"))
                continue

            # 输入拦截：命中直接回固定话术，不进模型
            rejected = check_input(user_content)
            if rejected:
                await ws.send_json(ServerMessage.token(rejected, 0))
                await ws.send_json(ServerMessage.done(1, remaining, rejected))
                await self.store.append_round(visitor_uuid, user_content, rejected)
                continue

            await self._stream_reply(ws, recv_queue, visitor_uuid, user_content)

    # ── 流式生成（可被 interrupt 打断）──

    async def _stream_reply(
        self,
        ws: WebSocket,
        recv_queue: asyncio.Queue,
        visitor_uuid: str,
        user_content: str,
    ) -> None:
        interrupted = False
        token_index = 0
        full_response = ""

        history = await self.store.get_history(visitor_uuid)
        stream = self.agent.astream(history, user_content)

        try:
            async for event in self._events_with_interrupt(ws, recv_queue, stream):
                if event["event"] == "token":
                    full_response += event["text"]
                    await ws.send_json(ServerMessage.token(event["text"], token_index))
                    token_index += 1
                elif event["event"] == "tool_call":
                    await ws.send_json(ServerMessage.tool_call(event["name"]))
        except _Interrupted:
            interrupted = True
        except Exception as e:
            logger.error("Agent stream 异常: %s", e, exc_info=True)
            await ws.send_json(
                ServerMessage.error("agent_error", "笔墨精灵思绪受阻，稍后再试。")
            )
            return

        final = filter_output(full_response)
        remaining = await self.store.remaining_today(visitor_uuid)
        await ws.send_json(ServerMessage.done(
            token_index, remaining, final, interrupted=interrupted))

        # 本轮入历史（打断的记已生成部分；存 final 保持记忆与展示一致）
        await self.store.append_round(visitor_uuid, user_content, final)

    async def _events_with_interrupt(self, ws, recv_queue, stream):
        """把 agent 事件流与接收队列复用：任何一方先就绪先处理。

        yield 正常事件；通过抛出 _Interrupted 表达打断，
        流正常结束则直接 return。
        """
        next_event = asyncio.ensure_future(anext(stream))
        try:
            while True:
                get_msg = asyncio.ensure_future(recv_queue.get())
                done, _ = await asyncio.wait(
                    {next_event, get_msg}, return_when=asyncio.FIRST_COMPLETED
                )

                if next_event in done:
                    try:
                        event = next_event.result()
                    except StopAsyncIteration:
                        # 流正常结束
                        if get_msg in done:
                            await self._handle_urgent(ws, get_msg.result())
                        else:
                            get_msg.cancel()
                        return
                    # 同批完成的接收消息优先处理（可能要求中断）
                    if get_msg in done:
                        if await self._handle_urgent(ws, get_msg.result()):
                            raise _Interrupted()
                    else:
                        get_msg.cancel()
                    yield event
                    next_event = asyncio.ensure_future(anext(stream))
                else:
                    # 只有接收侧就绪：ping 回应，interrupt 打断
                    if await self._handle_urgent(ws, get_msg.result()):
                        raise _Interrupted()
        finally:
            next_event.cancel()
            try:
                await stream.aclose()
            except Exception:
                pass

    @staticmethod
    async def _handle_urgent(ws: WebSocket, msg) -> bool:
        """处理流式期间收到的消息。返回 True 表示需要打断。"""
        if msg is None:
            return True  # 连接已断
        msg_type = msg.get("type", "") if isinstance(msg, dict) else ""
        if msg_type == "ping":
            await ws.send_json(ServerMessage.pong())
            return False
        if msg_type in ("interrupt", "message"):
            # 显式打断，或用户直接问了下一个问题
            return True
        return False


class _Interrupted(Exception):
    """内部信号：流式生成被访客打断"""
