package common

import (
	"strings"
	"testing"
)

func TestMakePasswordAndCheck(t *testing.T) {
	hash, err := MakePassword("s3cret!pwd")
	if err != nil {
		t.Fatalf("MakePassword 失败: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash 应为 argon2id 格式，got %q", hash[:min(20, len(hash))])
	}
	if !CheckPassword("s3cret!pwd", hash) {
		t.Error("正确密码校验应通过")
	}
	if CheckPassword("wrong-pwd", hash) {
		t.Error("错误密码校验不应通过")
	}
}

// 两次哈希应产生不同盐值
func TestMakePasswordUniqueSalt(t *testing.T) {
	h1, _ := MakePassword("same-pwd")
	h2, _ := MakePassword("same-pwd")
	if h1 == h2 {
		t.Error("同一密码两次哈希结果应不同（随机盐）")
	}
	if !CheckPassword("same-pwd", h1) || !CheckPassword("same-pwd", h2) {
		t.Error("两个哈希都应能校验原密码")
	}
}
