from guard.output_filter import filter_output


def test_phone_masked():
    out = filter_output("他的手机号是13812345678，别外传")
    assert "13812345678" not in out
    assert "***" in out


def test_email_masked():
    out = filter_output("发到 someone@example.com 吧")
    assert "someone@example.com" not in out
    assert "***@***" in out


def test_ai_prefix_rewritten():
    out = filter_output("作为一个AI助手，我觉得这首词意境极佳。")
    assert out.startswith("笔墨以为，")
    assert "AI" not in out.split("，")[0]


def test_clean_text_untouched():
    text = "落霞与孤鹜齐飞，秋水共长天一色。"
    assert filter_output(text) == text


def test_empty_safe():
    assert filter_output("") == ""
