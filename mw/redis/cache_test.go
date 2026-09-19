package redis

import (
	"testing"
	"time"
)

func TestBuildKey(t *testing.T) {
	cases := []struct {
		in   []interface{}
		want string
	}{
		{[]interface{}{"token", uint(42)}, "token:42"},
		{[]interface{}{"vcmtcnt", uint(7)}, "vcmtcnt:7"},
		{[]interface{}{"lock", "videos"}, "lock:videos"},
		{[]interface{}{int64(9), "suffix"}, "9:suffix"},
		{[]interface{}{int(3)}, "3"},
	}
	for _, c := range cases {
		if got := buildKey(c.in...); got != c.want {
			t.Errorf("buildKey(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestToInt64(t *testing.T) {
	if got := toInt64(uint(5)); got != 5 {
		t.Errorf("toInt64(uint) = %d", got)
	}
	if got := toInt64(int64(-3)); got != -3 {
		t.Errorf("toInt64(int64) = %d", got)
	}
	if got := toInt64(int(11)); got != 11 {
		t.Errorf("toInt64(int) = %d", got)
	}
	if got := toInt64("unsupported"); got != 0 {
		t.Errorf("不支持类型应返回 0, got %d", got)
	}
}

func TestKeyBuilders(t *testing.T) {
	if tokenK(1) != "token:1" {
		t.Errorf("tokenK = %q", tokenK(1))
	}
	if commentCountK(2) != "vcmtcnt:2" {
		t.Errorf("commentCountK = %q", commentCountK(2))
	}
	if userTotalFavK(3) != "utfav:3" {
		t.Errorf("userTotalFavK = %q", userTotalFavK(3))
	}
	if lockK("videos") != "lock:videos" {
		t.Errorf("lockK = %q", lockK("videos"))
	}
}

func TestParseUintMembers(t *testing.T) {
	got := parseUintMembers([]string{"1", "2", "3"})
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("parseUintMembers = %v", got)
	}
	if got := parseUintMembers(nil); len(got) != 0 {
		t.Errorf("空输入应返回空切片, got %v", got)
	}
	if got := parseUintMembers([]string{"1", "bad", "3"}); len(got) != 2 {
		t.Errorf("非法成员应被跳过, got %v", got)
	}
}

func TestJitterTTL(t *testing.T) {
	base := 10 * time.Second
	for i := 0; i < 200; i++ {
		got := jitterTTL(base)
		if got < base || got > base+base/5 {
			t.Fatalf("jitterTTL 超出 [base, base+20%%] 范围: %v", got)
		}
	}
}

func TestJitterTTLNonPositive(t *testing.T) {
	// RdbExpireTime 未初始化为 0，base<=0 时 fallback 到 0 → 返回 0 而不 panic
	if got := jitterTTL(0); got < 0 {
		t.Errorf("jitterTTL(0) 不应为负, got %v", got)
	}
}

func TestShortTTL(t *testing.T) {
	for i := 0; i < 50; i++ {
		got := shortTTL()
		if got < 30*time.Second || got > 36*time.Second {
			t.Fatalf("shortTTL 应在 30~36s 之间, got %v", got)
		}
	}
}
