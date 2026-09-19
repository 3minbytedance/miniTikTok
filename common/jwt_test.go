package common

import (
	"strings"
	"testing"
)

func TestGenerateAndParseToken(t *testing.T) {
	token := GenerateToken(12345, "alice")
	if token == "" || token == "fail" {
		t.Fatalf("GenerateToken 返回无效 token: %q", token)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken 解析失败: %v", err)
	}
	if claims.ID != 12345 {
		t.Errorf("claims.ID = %d, want 12345", claims.ID)
	}
	if claims.Username != "alice" {
		t.Errorf("claims.Username = %q, want %q", claims.Username, "alice")
	}
	if claims.Issuer != "DouShen" {
		t.Errorf("claims.Issuer = %q, want %q", claims.Issuer, "DouShen")
	}
}

func TestParseTokenInvalid(t *testing.T) {
	if _, err := ParseToken("not-a-jwt"); err == nil {
		t.Error("解析非法 token 应返回错误")
	}
}

func TestParseTokenTampered(t *testing.T) {
	token := GenerateToken(1, "bob")
	// 篡改签名部分最后一位
	tampered := token[:len(token)-1]
	if strings.HasSuffix(token, tampered[len(tampered)-1:]) {
		tampered += "A"
	}
	if _, err := ParseToken(tampered); err == nil {
		t.Error("解析被篡改的 token 应返回错误")
	}
}
