package common

import (
	"os"
	"strings"
	"testing"
)

// TestMain 将工作目录切到仓库根目录（词库按 ./common/ 相对路径加载）并初始化敏感词过滤器。
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	if err := InitSensitiveFilter(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestReplaceWordFilters(t *testing.T) {
	// 词库包含 sex / tmd 等词
	filtered := ReplaceWord("hello sex world tmd")
	if strings.Contains(filtered, "sex") || strings.Contains(filtered, "tmd") {
		t.Errorf("敏感词未被过滤: %q", filtered)
	}
	if strings.Count(filtered, "*") != 6 {
		t.Errorf("替换字符数 = %d, want 6 (%q)", strings.Count(filtered, "*"), filtered)
	}
}

func TestReplaceWordKeepsCleanText(t *testing.T) {
	clean := "今天天气不错，适合写代码 hello world 123"
	if got := ReplaceWord(clean); got != clean {
		t.Errorf("正常文本不应被替换: got %q, want %q", got, clean)
	}
}

func TestReplaceWordMixed(t *testing.T) {
	got := ReplaceWord("xgcdy")
	if got != "x***y" {
		t.Errorf("混排文本过滤结果 = %q, want %q", got, "x***y")
	}
}
