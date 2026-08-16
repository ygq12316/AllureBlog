"""多轮记忆与每日计数 — Redis 键控，按访客 uuid 隔离"""
import json
from datetime import datetime, timedelta, timezone

import redis.asyncio as aioredis

from config.settings import settings

# 受众为中文访客，自然日按东八区切分，服务器跑 UTC 也不错位
CN_TZ = timezone(timedelta(hours=8))


class ChatStore:
    """访客级对话存储：历史（LIST）+ 每日计数（按天键）"""

    def __init__(self, redis: aioredis.Redis | None = None):
        self._redis = redis
        self.ttl = settings.short_memory_ttl_seconds
        self.max_rounds = settings.max_history_rounds
        self.daily_limit = settings.daily_message_limit

    async def _get_redis(self) -> aioredis.Redis:
        if self._redis is None:
            self._redis = aioredis.from_url(settings.redis_url, decode_responses=True)
        return self._redis

    async def close(self):
        if self._redis:
            await self._redis.aclose()
            self._redis = None

    def _history_key(self, uuid: str) -> str:
        return f"chat:{uuid}:history"

    def _count_key(self, uuid: str) -> str:
        return f"chat:{uuid}:count:{datetime.now(CN_TZ).strftime('%Y%m%d')}"

    async def get_history(self, uuid: str) -> list[dict]:
        """读出最近 N 轮文本对话（user/assistant 交替）"""
        r = await self._get_redis()
        raw = await r.lrange(self._history_key(uuid), 0, -1) or []
        return [json.loads(h) for h in raw]

    async def append_round(self, uuid: str, user: str, assistant: str) -> None:
        """追加一轮对话，裁剪到记忆窗口并续期 TTL"""
        r = await self._get_redis()
        key = self._history_key(uuid)
        pipe = r.pipeline()
        pipe.rpush(key, json.dumps({"role": "user", "content": user}, ensure_ascii=False))
        pipe.rpush(key, json.dumps({"role": "assistant", "content": assistant}, ensure_ascii=False))
        pipe.ltrim(key, -(self.max_rounds * 2), -1)
        pipe.expire(key, self.ttl)
        await pipe.execute()

    async def incr_today(self, uuid: str) -> int:
        """当日计数 +1（首增设 48h 过期），返回自增后的值"""
        r = await self._get_redis()
        key = self._count_key(uuid)
        count = await r.incr(key)
        if count == 1:
            await r.expire(key, 48 * 3600)
        return count

    async def remaining_today(self, uuid: str) -> int:
        """当日剩余可用次数"""
        r = await self._get_redis()
        count = int(await r.get(self._count_key(uuid)) or 0)
        return max(0, self.daily_limit - count)
