package redis

import (
	"strconv"
	"strings"
)

// Delimiter redis key 分隔符
const Delimiter = ":"

func buildKey(parts ...interface{}) string {
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

// key 前缀
const (
	tokenKey        = "token"
	userNameKey     = "uname"  // uname:{uid} 用户名
	followSetKey    = "follow" // follow:{uid} 关注集合
	followerSetKey  = "follower"
	followCountKey  = "followcnt"
	followerCntKey  = "followercnt"
	feedKey         = "videos" // feed ZSet
	workCountKey    = "workcnt"
	commentCountKey = "vcmtcnt"
	videoFavCount   = "vfav"    // 视频点赞数
	userFavSet      = "ufavset" // 用户点赞视频集合
	userFavCount    = "ufavcnt" // 用户点赞数
	userTotalFav    = "utfav"   // 作者获赞总数
	lockPrefix      = "lock"
)

func tokenK(uid uint) string     { return buildKey(tokenKey, uid) }
func userNameK(uid uint) string  { return buildKey(userNameKey, uid) }
func followSetK(uid uint) string { return buildKey(followSetKey, uid) }
func followerSetK(uid uint) string {
	return buildKey(followerSetKey, uid)
}
func followCountK(uid uint) string  { return buildKey(followCountKey, uid) }
func followerCntK(uid uint) string  { return buildKey(followerCntKey, uid) }
func workCountK(uid uint) string    { return buildKey(workCountKey, uid) }
func commentCountK(vid uint) string { return buildKey(commentCountKey, vid) }
func videoFavK(vid uint) string     { return buildKey(videoFavCount, vid) }
func userFavSetK(uid uint) string   { return buildKey(userFavSet, uid) }
func userFavCntK(uid uint) string   { return buildKey(userFavCount, uid) }
func userTotalFavK(uid uint) string { return buildKey(userTotalFav, uid) }
func lockK(key string) string       { return buildKey(lockPrefix, key) }
