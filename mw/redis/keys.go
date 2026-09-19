package redis

import (
	"strconv"
	"strings"
)

// Delimiter redis key 分隔符
const Delimiter = ":"

// BuildKey 用分隔符拼接 redis key 片段。
func BuildKey(parts ...interface{}) string {
	ss := make([]string, 0, len(parts))
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			ss = append(ss, v)
		case uint:
			ss = append(ss, strconv.FormatUint(uint64(v), 10))
		case int64:
			ss = append(ss, strconv.FormatInt(v, 10))
		case int:
			ss = append(ss, strconv.Itoa(v))
		default:
			ss = append(ss, strconv.FormatInt(int64(toInt64(p)), 10))
		}
	}
	return strings.Join(ss, Delimiter)
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case uint:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

// mw/redis 仅保留网关与各服务共享的基础设施 key；
// 各业务域 key 定义在对应服务内（service/social/cache.go、service/video/cache.go）。
const (
	tokenKey   = "token"
	lockPrefix = "lock"
)

func tokenK(uid uint) string { return BuildKey(tokenKey, uid) }
func lockK(key string) string {
	return BuildKey(lockPrefix, key)
}
