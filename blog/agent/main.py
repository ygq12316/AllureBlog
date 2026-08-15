# blog/agent/main.py
import asyncio
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, WebSocket

from config.settings import settings
from memory.short_term import ShortTermMemory
from agent.core import create_bi_mo_agent
from ws.handler import ChatHandler

# 配置日志
logging.basicConfig(
    level=getattr(logging, settings.log_level.upper(), logging.INFO),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)

# ── 全局服务实例 ──
short_memory: ShortTermMemory | None = None
agent = None
chat_handler: ChatHandler | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    global short_memory, agent, chat_handler

    logger.info("笔墨精灵正在苏醒...")

    # 初始化服务
    short_memory = ShortTermMemory()

    # 创建 Agent
    agent = await create_bi_mo_agent()

    # 创建 WebSocket handler
    chat_handler = ChatHandler(
        agent=agent,
        short_term_memory=short_memory,
    )

    logger.info("笔墨精灵已就绪。")

    yield

    # 清理
    logger.info("笔墨精灵正在休憩...")
    await short_memory.close()


app = FastAPI(
    title="笔墨精灵智能体服务",
    description="博客「笔墨 · Ink & Code」的访客互动伴侣",
    version="0.1.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health():
    """健康检查"""
    return {
        "status": "ok",
        "service": "笔墨精灵智能体",
        "version": "0.1.0",
    }


@app.websocket("/chat/ws")
async def chat_websocket(ws: WebSocket):
    """WebSocket 聊天端点 — 笔墨精灵流式对话"""
    if not chat_handler:
        await ws.accept()
        await ws.send_json({
            "type": "error",
            "code": "not_ready",
            "display": "笔墨精灵尚在苏醒中，稍候再来。",
        })
        await ws.close()
        return

    await chat_handler.handle(ws)