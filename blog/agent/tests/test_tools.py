import httpx
import pytest

from agent.tools import TOOLS_SPEC, dispatch_tool, search_blog
from config import settings as settings_mod

GO_ARTICLES = {"articles": [
    {"title": "Vue3 组合式 API 实战", "category": "前端"},
    {"title": "Go 中的 SQLite 实践", "category": "后端"},
], "total": 2}
GO_NOTES = {"notes": [{"title": "深夜写代码的仪式感"}], "total": 1}


def fake_transport(request: httpx.Request) -> httpx.Response:
    if request.url.path == "/api/articles":
        return httpx.Response(200, json=GO_ARTICLES)
    if request.url.path == "/api/notes":
        return httpx.Response(200, json=GO_NOTES)
    return httpx.Response(404, json={"error": "not found"})


@pytest.fixture
def fake_client():
    return httpx.AsyncClient(transport=httpx.MockTransport(fake_transport), timeout=3)


@pytest.fixture(autouse=True)
def fake_base(monkeypatch):
    monkeypatch.setattr(settings_mod.settings, "blog_api_base", "http://go-test")


async def test_search_all_titles(fake_client):
    out = await search_blog("", client=fake_client)
    assert "文章《Vue3 组合式 API 实战》 分类:前端" in out
    assert "文章《Go 中的 SQLite 实践》 分类:后端" in out
    assert "随笔《深夜写代码的仪式感》" in out
    assert "（共 3 条）" in out


async def test_search_with_keyword_case_insensitive(fake_client):
    out = await search_blog("vue3", client=fake_client)
    assert "Vue3" in out
    assert "SQLite" not in out


async def test_search_no_match(fake_client):
    out = await search_blog("量子力学", client=fake_client)
    assert "未找到" in out


async def test_search_api_down_returns_apology():
    # 指向必然连接失败的端口，验证降级话术而非抛异常
    from config import settings as s
    s.settings.blog_api_base = "http://127.0.0.1:1"
    out = await search_blog("")
    assert "检索暂时不可用" in out


async def test_dispatch_tool_routes(fake_client, monkeypatch):
    async def fake_search(keyword="", client=None):
        return "检索结果"
    monkeypatch.setattr("agent.tools.search_blog", fake_search)
    out = await dispatch_tool("search_blog", '{"keyword": "诗"}')
    assert out == "检索结果"
    assert await dispatch_tool("unknown_tool", "{}") == "未知工具: unknown_tool"


def test_tools_spec_shape():
    spec = TOOLS_SPEC[0]["function"]
    assert spec["name"] == "search_blog"
    assert spec["parameters"]["type"] == "object"
