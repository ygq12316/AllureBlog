import fakeredis.aioredis

from memory.chat_store import ChatStore


async def make_store():
    redis = fakeredis.aioredis.FakeRedis(decode_responses=True)
    return ChatStore(redis=redis), redis


async def test_append_and_get_history():
    store, redis = await make_store()
    await store.append_round("u1", "你好", "幸会")
    history = await store.get_history("u1")
    assert history == [
        {"role": "user", "content": "你好"},
        {"role": "assistant", "content": "幸会"},
    ]
    await redis.aclose()


async def test_history_window_trims_to_5_rounds():
    store, redis = await make_store()
    for i in range(8):  # 8 轮，窗口只留最近 5 轮
        await store.append_round("u1", f"问{i}", f"答{i}")
    history = await store.get_history("u1")
    assert len(history) == 10
    assert history[0] == {"role": "user", "content": "问3"}
    assert history[-1] == {"role": "assistant", "content": "答7"}
    await redis.aclose()


async def test_incr_and_remaining():
    store, redis = await make_store()
    for i in range(10):
        count = await store.incr_today("u1")
        assert count == i + 1
    assert await store.remaining_today("u1") == 0
    # 第 11 次：INCR 仍自增（handler 据此判定超限）
    assert await store.incr_today("u1") == 11
    await redis.aclose()


async def test_remaining_fresh_visitor():
    store, redis = await make_store()
    assert await store.remaining_today("u2") == 10
    await redis.aclose()


async def test_history_isolated_per_visitor():
    store, redis = await make_store()
    await store.append_round("u1", "a", "b")
    assert await store.get_history("u2") == []
    await redis.aclose()
