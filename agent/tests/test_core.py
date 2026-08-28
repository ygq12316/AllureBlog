from types import SimpleNamespace as NS

from agent.core import SYSTEM_PROMPT, BiMoAgent


class FakeStream:
    """openai 流式响应替身：按序吐 chunk"""

    def __init__(self, chunks):
        self._chunks = list(chunks)

    def __aiter__(self):
        return self

    async def __anext__(self):
        if not self._chunks:
            raise StopAsyncIteration
        return self._chunks.pop(0)


def text_chunk(text):
    return NS(choices=[NS(delta=NS(content=text, tool_calls=None), finish_reason=None)])


def tool_chunk(index, tc_id=None, name=None, arguments=None):
    fn = NS(name=name, arguments=arguments)
    return NS(choices=[NS(delta=NS(content=None,
                                    tool_calls=[NS(index=index, id=tc_id, function=fn)]),
                               finish_reason=None)])


class FakeCompletions:
    """按脚本依次返回假流，并记录每次 create 的参数"""

    def __init__(self, script):
        self.script = list(script)
        self.calls = []

    async def create(self, **kw):
        self.calls.append(kw)
        return self.script.pop(0)


class FakeClient:
    def __init__(self, script):
        self.chat = NS(completions=FakeCompletions(script))


async def collect(agent, history, msg):
    return [e async for e in agent.astream(history, msg)]


async def test_plain_text_stream():
    client = FakeClient([FakeStream([text_chunk("你好"), text_chunk("呀")])])
    agent = BiMoAgent(client=client)
    events = await collect(agent, [], "打招呼")
    assert events == [
        {"event": "token", "text": "你好"},
        {"event": "token", "text": "呀"},
    ]
    # 第一轮允许工具
    assert client.chat.completions.calls[0]["tools"] is not None
    # 请求里带 system prompt 与用户消息
    sent = client.chat.completions.calls[0]["messages"]
    assert sent[0]["content"] == SYSTEM_PROMPT
    assert sent[-1] == {"role": "user", "content": "打招呼"}


async def test_history_passed_through():
    client = FakeClient([FakeStream([text_chunk("嗯")])])
    agent = BiMoAgent(client=client)
    history = [
        {"role": "user", "content": "之前问过"},
        {"role": "assistant", "content": "之前答过"},
    ]
    await collect(agent, history, "新问题")
    sent = client.chat.completions.calls[0]["messages"]
    assert sent[1:3] == history  # system 之后、本条之前


async def test_tool_call_incremental_assembly(monkeypatch):
    # 工具调用参数被 openai 流式分片推送：name 一次到位，arguments 两段拼接
    async def fake_dispatch(name, args_json):
        fake_dispatch.called_with = (name, args_json)
        return "检索结果"
    fake_dispatch.called_with = None
    from agent import tools as tools_mod
    monkeypatch.setattr(tools_mod, "dispatch_tool", fake_dispatch)

    script = [
        FakeStream([
            tool_chunk(0, tc_id="call_1", name="search_blog"),
            tool_chunk(0, arguments='{"key'),
            tool_chunk(0, arguments='word": "诗"}'),
        ]),
        FakeStream([text_chunk("共觅得两篇")]),
    ]
    client = FakeClient(script)
    agent = BiMoAgent(client=client)
    events = await collect(agent, [], "博客里有关于诗的文章吗")
    assert {"event": "tool_call", "name": "search_blog", "args": '{"keyword": "诗"}'} in events
    assert {"event": "token", "text": "共觅得两篇"} in events
    assert fake_dispatch.called_with == ("search_blog", '{"keyword": "诗"}')
    # 第二轮请求包含 assistant(tool_calls) 与 tool 结果
    sent2 = client.chat.completions.calls[1]["messages"]
    roles = [m["role"] for m in sent2]
    assert "tool" in roles
    assert roles[-2] == "assistant"


async def test_tool_loop_ceiling(monkeypatch):
    # 连续三轮都要求工具：第三轮 create 不再传 tools，循环终止
    dispatch_count = {"n": 0}

    async def fake_dispatch(name, args_json):
        dispatch_count["n"] += 1
        return "结果"

    from agent import tools as tools_mod
    monkeypatch.setattr(tools_mod, "dispatch_tool", fake_dispatch)

    script = [
        FakeStream([tool_chunk(0, tc_id="c1", name="search_blog")]),
        FakeStream([tool_chunk(0, tc_id="c2", name="search_blog")]),
        # 第三轮请求不带 tools，真实模型只能文字作答 — fake 忠实模拟该行为
        FakeStream([text_chunk("最终回答")]),
    ]
    client = FakeClient(script)
    agent = BiMoAgent(client=client)
    events = await collect(agent, [], "连续检索")
    # 第三轮（attempt=2）allow_tools=False：请求无 tools 参数，模型只能文字作答
    third_call = client.chat.completions.calls[2]
    assert "tools" not in third_call or third_call["tools"] is None
    # 工具至多执行 2 次
    assert dispatch_count["n"] == 2
    # 事件里只有前两次 tool_call，第三轮以文本收尾
    assert sum(1 for e in events if e["event"] == "tool_call") == 2
    assert {"event": "token", "text": "最终回答"} in events
