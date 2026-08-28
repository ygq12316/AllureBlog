"""检索工具 — search_blog：从 Go 后端拉取文章/随笔标题"""
import asyncio
import json
import logging

import httpx

from config.settings import settings

logger = logging.getLogger(__name__)

TOOLS_SPEC = [
    {
        "type": "function",
        "function": {
            "name": "search_blog",
            "description": "检索博客内的文章与随笔标题。访客想了解博客写了什么、找某主题的文章/随笔时调用。",
            "parameters": {
                "type": "object",
                "properties": {
                    "keyword": {"type": "string", "description": "标题过滤关键词；留空返回全部"},
                },
                "required": [],
            },
        },
    }
]


async def search_blog(keyword: str = "", client: httpx.AsyncClient | None = None) -> str:
    """并行拉取文章与随笔标题，按关键词过滤，格式化为文本。

    Go API 不可达时返回致歉话术（由模型转述给访客），不抛异常。
    """
    kw = (keyword or "").strip().lower()
    own = client is None
    if client is None:
        client = httpx.AsyncClient(timeout=3)
    try:
        r_articles, r_notes = await asyncio.gather(
            client.get(f"{settings.blog_api_base}/api/articles", params={"page": 1, "per_page": 50}),
            client.get(f"{settings.blog_api_base}/api/notes", params={"page": 1, "per_page": 50}),
        )
        articles = r_articles.json().get("articles") or [] if r_articles.status_code == 200 else []
        notes = r_notes.json().get("notes") or [] if r_notes.status_code == 200 else []
    except httpx.HTTPError as e:
        logger.warning("[search_blog] 博客 API 不可达: %s", e)
        return "检索暂时不可用，请向访客致歉并建议稍后再问。"
    finally:
        if own:
            await client.aclose()

    lines = []
    for a in articles:
        title = str(a.get("title") or "")
        if kw and kw not in title.lower():
            continue
        category = a.get("category") or "未分类"
        lines.append(f"文章《{title}》 分类:{category}")
    for n in notes:
        title = str(n.get("title") or "")
        if kw and kw not in title.lower():
            continue
        lines.append(f"随笔《{title}》")

    if not lines:
        return f"未找到与「{keyword}」相关的内容。"

    total = len(lines)
    shown = lines[:20]
    shown.append(f"（共 {total} 条，仅列前 20 条）" if total > 20 else f"（共 {total} 条）")
    return "\n".join(shown)


async def dispatch_tool(name: str, arguments_json: str) -> str:
    """执行模型请求的工具调用，返回给模型的文本结果"""
    if name == "search_blog":
        try:
            args = json.loads(arguments_json or "{}")
        except json.JSONDecodeError:
            args = {}
        return await search_blog(str(args.get("keyword") or ""))
    return f"未知工具: {name}"
