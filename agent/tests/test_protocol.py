from ws.protocol import ClientMessage, ServerMessage


def test_auth_result_with_remaining():
    msg = ServerMessage.auth_result(True, "幸会", remaining=10)
    assert msg == {"type": "auth_result", "success": True, "greeting": "幸会", "remaining": 10}


def test_auth_result_failure_with_code():
    msg = ServerMessage.auth_result(False, "请先登录", code="login_required")
    assert msg["success"] is False
    assert msg["code"] == "login_required"
    assert "remaining" not in msg


def test_tool_call_message():
    assert ServerMessage.tool_call("search_blog") == {"type": "tool_call", "name": "search_blog"}


def test_done_with_final_and_remaining():
    msg = ServerMessage.done(total_tokens=5, remaining=3, final="定稿文本", interrupted=True)
    assert msg == {
        "type": "done", "total_tokens": 5, "remaining": 3,
        "final": "定稿文本", "interrupted": True,
    }


def test_error_shape():
    assert ServerMessage.error("limit_exceeded", "今日笔墨已尽") == {
        "type": "error", "code": "limit_exceeded", "display": "今日笔墨已尽",
    }


def test_client_message_parse():
    m = ClientMessage.model_validate({"type": "auth", "visitor_uuid": "u1"})
    assert m.type == "auth" and m.visitor_uuid == "u1"
