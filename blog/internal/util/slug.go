package util

import (
	"math/rand"
	"regexp"
	"strings"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9一-鿿]+`)

func Slugify(title string) string {
	title = strings.ToLower(title)
	title = strings.TrimSpace(title)
	slug := nonAlphaNum.ReplaceAllString(title, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "untitled"
	}
	return slug
}

func RandomSuffix(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
