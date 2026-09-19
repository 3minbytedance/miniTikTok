package common

import (
	"errors"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestGetCurrentUserID(t *testing.T) {
	rc := &app.RequestContext{}
	rc.Set(ContextUserIDKey, uint(42))

	uid, err := GetCurrentUserID(rc)
	if err != nil {
		t.Fatalf("GetCurrentUserID 失败: %v", err)
	}
	if uid != 42 {
		t.Errorf("uid = %d, want 42", uid)
	}
}

func TestGetCurrentUserIDNotLogin(t *testing.T) {
	rc := &app.RequestContext{}
	_, err := GetCurrentUserID(rc)
	if !errors.Is(err, ErrorUserNotLogin) {
		t.Errorf("未登录应返回 ErrorUserNotLogin, got %v", err)
	}
}

func TestGetCurrentUserIDWrongType(t *testing.T) {
	rc := &app.RequestContext{}
	rc.Set(ContextUserIDKey, "not-a-uint")
	if _, err := GetCurrentUserID(rc); err == nil {
		t.Error("类型断言失败应返回错误")
	}
}
