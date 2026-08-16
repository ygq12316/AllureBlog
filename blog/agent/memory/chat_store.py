"""多轮记忆与每日计数 — 进程内存储，按访客 uuid 隔离

单进程单 worker（uvicorn 默认）运行，事件循环内的读写无 await 间隙，
天然原子；历史窗口用 deque 环形裁剪，TTL 与日计数按墙钟惰性判定，
无需外部 Redis。
"""
import time
from collections import deque
from collections.abc import Callable
from datetime import datetime, timedelta, timezone

from config.settings import settings

# 受众为中文访客，自然日按东八区切分，服务器跑 UTC 也不错位
CN_TZ = timezone(timedelta(hours=8))


class ChatStore:
    """访客级对话存储：历史（环形窗口 + TTL）+ 东八区日计数"""

    def __init__(self, *, clock: Callable[[], float] = time.monotonic, wall: Callable[[], datetime] | None = None):
        self.ttl = settings.short_memory_ttl_seconds
        self.max_rounds = settings.max_history_rounds
        self.daily_limit = settings.daily_message_limit
        # uuid -> {"rounds": deque[(user, assistant)], "expire_at": 单调秒}
        self._history: dict[str, dict] = {}
        # uuid -> (YYYYMMDD, 当日已用条数)
        self._counts: dict[str, tuple[str, int]] = {}
        self._clock = clock
        self._wall = wall or datetime.now

    def close(self) -> None:
        """清空内存（无外部资源，供 lifespan 关闭时调用）"""
        self._history.clear()
        self._counts.clear()

    def _today(self) -> str:
        return self._wall().astimezone(CN_TZ).strftime("%Y%m%d")

    def _live_rounds(self, uuid: str) -> deque | None:
        entry = self._history.get(uuid)
        if entry is None:
            return None
        if entry["expire_at"] <= self._clock():
            del self._history[uuid]
            return None
        return entry["rounds"]

    async def get_history(self, uuid: str) -> list[dict]:
        """读出最近 N 轮文本对话（user/assistant 交替）"""
        rounds = self._live_rounds(uuid)
        if not rounds:
            return []
        history: list[dict] = []
        for user, assistant in rounds:
            history.append({"role": "user", "content": user})
            history.append({"role": "assistant", "content": assistant})
        return history

    async def append_round(self, uuid: str, user: str, assistant: str) -> None:
        """追加一轮对话，裁剪到记忆窗口并续期 TTL"""
        rounds = self._live_rounds(uuid)
        if rounds is None:
            rounds = deque()
            # 占位永不过期，末尾写入真实到期时刻，避免中途清扫误删
            self._history[uuid] = {"rounds": rounds, "expire_at": float("inf")}
        rounds.append((user, assistant))
        while len(rounds) > self.max_rounds:
            rounds.popleft()
        self._history[uuid]["expire_at"] = self._clock() + self.ttl
        # 顺带清扫过期残留与隔日计数，防长期运行内存泄漏
        now = self._clock()
        for key in [k for k, e in self._history.items() if e["expire_at"] <= now]:
            del self._history[key]
        today = self._today()
        self._counts = {k: v for k, v in self._counts.items() if v[0] == today}

    async def incr_today(self, uuid: str) -> int:
        """当日计数 +1，返回自增后的值（跨日自然归零）"""
        today = self._today()
        entry = self._counts.get(uuid)
        count = entry[1] + 1 if entry and entry[0] == today else 1
        self._counts[uuid] = (today, count)
        return count

    async def remaining_today(self, uuid: str) -> int:
        """当日剩余可用次数"""
        entry = self._counts.get(uuid)
        count = entry[1] if entry and entry[0] == self._today() else 0
        return max(0, self.daily_limit - count)
