# blog/agent/ws/handler.py
"""WebSocket 聊天处理器 — 管理 DeepAgents 流式对话的生命周期"""
import asyncio
import logging

from fastapi import WebSocket, WebSocketDisconnect

from ws.protocol import ClientMessage, ServerMessage

logger = logging.getLogger(__name__)


class ChatHandler:
    """WebSocket 聊天处理器 — 管理连接生命周期和流式推送"""

    def __init__(self, agent, short_term_memory):
        self.agent = agent
        self.short_term_memory = short_term_memory

    async def handle(self, ws: WebSocket) -> None:
        """主处理入口"""
        await ws.accept()

        # ── Step 1: 等待并校验 auth 消息 ──
        try:
            raw = await ws.receive_json()
        except Exception:
            await self._reject(ws, "请先完成身份认证")
            return

        try:
            auth = ClientMessage.model_validate(raw)
        except Exception:
            await self._reject(ws, "消息格式错误")
            return

        # visitor_uuid 缺失会让所有匿名访客共享同一会话，必须拒绝
        if auth.type != "auth" or not (auth.visitor_uuid or "").strip():
            await self._reject(ws, "请先完成身份认证")
            return

        visitor_uuid = auth.visitor_uuid.strip()
        visitor_name = (auth.visitor_name or "访客").strip() or "访客"

        # ── Step 2: 个性化问候 ──
        greeting = f"笔墨精灵在此，愿与你共话天地，{visitor_name}。"
        await ws.send_json(ServerMessage.auth_result(True, greeting))

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

    @staticmethod
    async def _reject(ws: WebSocket, display: str) -> None:
        await ws.send_json(ServerMessage.error("auth_required", display))
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

        config = {
            "configurable": {
                "thread_id": visitor_uuid,
                "visitor_uuid": visitor_uuid,
            },
        }

        stream = self.agent.astream_events(
            {"messages": [{"role": "user", "content": user_content}]},
            config=config,
            version="v2",
        )

        try:
            async for event in self._events_with_interrupt(ws, recv_queue, stream):
                token = self._extract_token(event)
                if token:
                    full_response += token
                    await ws.send_json(ServerMessage.token(token, token_index))
                    token_index += 1
        except _Interrupted:
            interrupted = True
        except Exception as e:
            logger.error("Agent stream 异常: %s", e, exc_info=True)
            await ws.send_json(
                ServerMessage.error("agent_error", "笔墨精灵思绪受阻，稍后再试。")
            )
            return

        await ws.send_json(ServerMessage.done(token_index, [], interrupted=interrupted))

        # 对话轮次落 Redis（跨重启的短期记忆审计；会话上下文由 thread_id 管理）
        await self._remember(visitor_uuid, "user", user_content)
        if full_response:
            await self._remember(visitor_uuid, "assistant", full_response)

    async def _events_with_interrupt(self, ws: WebSocket, recv_queue: asyncio.Queue, stream):
        """把流事件与接收队列复用：任何一方先就绪先处理。

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

    @staticmethod
    def _extract_token(event: dict) -> str | None:
        if event.get("event") != "on_chat_model_stream":
            return None
        chunk = event.get("data", {}).get("chunk")
        if chunk and hasattr(chunk, "content") and chunk.content:
            token = chunk.content
            if isinstance(token, str) and token:
                return token
        return None

    async def _remember(self, visitor_uuid: str, role: str, content: str) -> None:
        try:
            await self.short_term_memory.append(visitor_uuid, role, content)
        except Exception as e:
            logger.warning("短期记忆写入失败: %s", e)


class _Interrupted(Exception):
    """内部信号：流式生成被访客打断"""
