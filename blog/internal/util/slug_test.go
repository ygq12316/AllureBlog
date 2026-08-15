package util

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"英文转小写", "Hello World", "hello-world"},
		{"保留中文", "笔墨精灵：水墨画", "笔墨精灵-水墨画"},
		{"多分隔符合一", "Go  &&  Gin // Gorm", "go-gin-gorm"},
		{"去除首尾连字符", "--标题--", "标题"},
		{"全特殊字符回退", "!!!???***", "untitled"},
		{"空字符串回退", "", "untitled"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Slugify(tc.input); got != tc.want {
				t.Errorf("Slugify(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRandomSuffix(t *testing.T) {
	for _, n := range []int{1, 6, 12} {
		if got := RandomSuffix(n); len(got) != n {
			t.Errorf("RandomSuffix(%d) 长度 = %d, want %d", n, len(got), n)
		}
	}
	// 两次生成应不同（碰撞概率可忽略）
	if RandomSuffix(8) == RandomSuffix(8) {
		t.Error("RandomSuffix 连续两次生成相同值")
	}
}
