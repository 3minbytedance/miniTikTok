package common

import "testing"

func TestMapErrMsg(t *testing.T) {
	if got := MapErrMsg(CodeSuccess); got != "success" {
		t.Errorf("MapErrMsg(0) = %q, want %q", got, "success")
	}
	if got := MapErrMsg(CodeLimiterCount); got != "请求次数过多，已被限制，稍后再试" {
		t.Errorf("限流文案不符: %q", got)
	}
	// 未定义错误码返回兜底文案
	if got := MapErrMsg(9999); got == "" || got == "success" {
		t.Errorf("未定义错误码应返回兜底文案: %q", got)
	}
}

func TestIsCodeErr(t *testing.T) {
	if !IsCodeErr(CodeInvalidParam) {
		t.Error("已定义错误码应返回 true")
	}
	if IsCodeErr(9999) {
		t.Error("未定义错误码应返回 false")
	}
}
