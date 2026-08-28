package util

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string // 输出应包含的片段
	}{
		{"标题", "# 笔墨", "<h1"},
		{"粗体", "**重点**", "<strong>重点</strong>"},
		{"代码块", "```go\nfmt.Println(1)\n```", "<pre"},
		{"换行", "第一行\n第二行", "第一行"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderMarkdown(tc.input)
			if !strings.Contains(got, tc.want) {
				t.Errorf("RenderMarkdown(%q) = %q, 应包含 %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestRenderMarkdownEscapesRawHTML 验证未开启 WithUnsafe 时，
// Markdown 中内嵌的原始 HTML 不会原样透传（存储型 XSS 防线）
func TestRenderMarkdownEscapesRawHTML(t *testing.T) {
	got := RenderMarkdown("<script>alert(1)</script>正文")
	if strings.Contains(got, "<script>") {
		t.Errorf("原始 <script> 标签透传到了输出: %q", got)
	}
	got2 := RenderMarkdown("<img src=x onerror=alert(1)>")
	if strings.Contains(got2, "onerror=") {
		t.Errorf("原始 <img onerror> 透传到了输出: %q", got2)
	}
}

func TestExtractExcerpt(t *testing.T) {
	html := "<p>" + strings.Repeat("墨", 200) + "</p>"
	got := ExtractExcerpt(html, 120)
	if got != strings.Repeat("墨", 120)+"..." {
		t.Errorf("超长摘要未按 120 截断: 长度 %d", len([]rune(got)))
	}
	if got := ExtractExcerpt("<p>短文</p>", 120); got != "短文" {
		t.Errorf("短文本被错误修改: %q", got)
	}
}
