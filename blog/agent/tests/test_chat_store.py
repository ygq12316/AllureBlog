"""ChatStore 进程内存实现 — TTL / 窗口 / 日计数的纯行为测试"""
from datetime import datetime

from memory.chat_store import CN_TZ, ChatStore


class FakeClock:
    """单调时钟桩，可手动推进"""

    def __init__(self):
        self.now = 0.0

    def __call__(self):
        return self.now

    def advance(self, seconds: float):
        self.now += seconds


def make_store(wall=None):
    clock = FakeClock()
    wall = wall or (lambda: datetime(2026, 8, 16, 12, 0, tzinfo=CN_TZ))
    return ChatStore(clock=clock, wall=wall), clock


async def test_append_and_get_history():
    store, _ = make_store()
    await store.append_round("u1", "你好", "幸会")
    assert await store.get_history("u1") == [
        {"role": "user", "content": "你好"},
        {"role": "assistant", "content": "幸会"},
    ]


async def test_history_window_trims_to_5_rounds():
    store, _ = make_store()
    for i in range(8):  # 8 轮，窗口只留最近 5 轮
        await store.append_round("u1", f"问{i}", f"答{i}")
    history = await store.get_history("u1")
    assert len(history) == 10
    assert history[0] == {"role": "user", "content": "问3"}
    assert history[-1] == {"role": "assistant", "content": "答7"}


async def test_history_expires_after_ttl():
    store, clock = make_store()
    await store.append_round("u1", "你好", "幸会")
    clock.advance(store.ttl + 1)
    assert await store.get_history("u1") == []


async def test_ttl_refreshed_on_each_round():
    store, clock = make_store()
    await store.append_round("u1", "问1", "答1")
    clock.advance(store.ttl - 10)
    await store.append_round("u1", "问2", "答2")
    clock.advance(store.ttl - 10)  # 第一轮的 TTL 早已过，但已被第二轮续期
    history = await store.get_history("u1")
    assert len(history) == 4


async def test_incr_and_remaining():
    store, _ = make_store()
    for i in range(10):
        assert await store.incr_today("u1") == i + 1
    assert await store.remaining_today("u1") == 0
    # 第 11 次：仍自增（handler 据此判定超限）
    assert await store.incr_today("u1") == 11


async def test_remaining_fresh_visitor():
    store, _ = make_store()
    assert await store.remaining_today("u2") == 10


async def test_history_isolated_per_visitor():
    store, _ = make_store()
    await store.append_round("u1", "a", "b")
    assert await store.get_history("u2") == []


async def test_count_resets_next_day():
    now = {"t": datetime(2026, 8, 16, 23, 50, tzinfo=CN_TZ)}
    store, _ = make_store(wall=lambda: now["t"])
    await store.incr_today("u1")
    await store.incr_today("u1")
    # 跨过东八区午夜：计数自然归零
    now["t"] = datetime(2026, 8, 17, 0, 10, tzinfo=CN_TZ)
    assert await store.remaining_today("u1") == 10
    assert await store.incr_today("u1") == 1
