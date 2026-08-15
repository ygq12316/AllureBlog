package util

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

func RenderMarkdown(input string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		return input
	}
	return buf.String()
}

func ExtractExcerpt(html string, maxLen int) string {
	plain := stripTags(html)
	runes := []rune(plain)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return plain
}

func stripTags(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(b.String())
}
