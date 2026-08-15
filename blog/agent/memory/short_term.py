# blog/agent/memory/short_term.py
import json
import redis.asyncio as aioredis
from config.settings import settings


class ShortTermMemory:
    """短期记忆 — 基于 Redis 的会话上下文"""

    def __init__(self, redis_url: str | None = None):
        self.redis_url = redis_url or settings.redis_url
        self.ttl = settings.short_memory_ttl_seconds
        self.max_history = settings.max_history_rounds
        self._redis: aioredis.Redis | None = None

    async def _get_redis(self) -> aioredis.Redis:
        if self._redis is None:
            self._redis = aioredis.from_url(self.redis_url, decode_responses=True)
        return self._redis

    async def close(self):
        if self._redis:
            await self._redis.aclose()
            self._redis = None

    def _keys(self, visitor_uuid: str) -> dict[str, str]:
        return {
            "context": f"chat:{visitor_uuid}:context",
            "history": f"chat:{visitor_uuid}:history",
            "state": f"chat:{visitor_uuid}:state",
        }

    async def load(self, visitor_uuid: str) -> dict:
        """加载访客的短期记忆"""
        r = await self._get_redis()
        keys = self._keys(visitor_uuid)

        history_raw = await r.lrange(keys["history"], 0, -1) or []
        context_raw = await r.hgetall(keys["context"]) or {}
        state_raw = await r.get(keys["state"]) or "{}"

        return {
            "history": [json.loads(h) for h in history_raw],
            "context": context_raw,
            "state": json.loads(state_raw),
        }

    async def append(self, visitor_uuid: str, role: str, content: str) -> None:
        """追加一轮对话到历史"""
        r = await self._get_redis()
        keys = self._keys(visitor_uuid)

        entry = json.dumps({"role": role, "content": content})
        pipe = r.pipeline()
        pipe.rpush(keys["history"], entry)
        # 保持不超过 max_history 轮（每轮2条：user+assistant）
        pipe.ltrim(keys["history"], -(self.max_history * 2), -1)
        pipe.expire(keys["history"], self.ttl)
        await pipe.execute()

    async def get_context(self, visitor_uuid: str) -> dict:
        """获取当前对话上下文"""
        r = await self._get_redis()
        keys = self._keys(visitor_uuid)
        raw = await r.hgetall(keys["context"])
        return raw or {}

    async def update_context(self, visitor_uuid: str, context: dict) -> None:
        """更新对话上下文"""
        r = await self._get_redis()
        keys = self._keys(visitor_uuid)
        if context:
            pipe = r.pipeline()
            pipe.delete(keys["context"])
            pipe.hset(keys["context"], mapping=context)
            pipe.expire(keys["context"], self.ttl)
            await pipe.execute()

    async def persist_summary(self, visitor_uuid: str) -> str | None:
        """将当前对话历史提炼为摘要，返回摘要文本"""
        r = await self._get_redis()
        keys = self._keys(visitor_uuid)
        history_raw = await r.lrange(keys["history"], 0, -1)
        if not history_raw:
            return None

        # 简单合并最近对话作为摘要（实际由 Deepagents memory skill 完成）
        messages = [json.loads(h) for h in history_raw]
        summary_parts = []
        for m in messages[-4:]:  # 取最近4条
            role = "访客" if m["role"] == "user" else "精灵"
            summary_parts.append(f"{role}: {m['content'][:80]}")

        return "；".join(summary_parts)

    async def clear(self, visitor_uuid: str) -> None:
        """清除短期记忆"""
        r = await self._get_redis()
        keys = self._keys(visitor_uuid)
        await r.delete(keys["context"], keys["history"], keys["state"])
