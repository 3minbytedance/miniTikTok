package common

import (
	"errors"
	"github.com/cloudwego/hertz/pkg/app"
)

const ContextUserIDKey = "user_id"
const TokenValid = "tokenValid"

var ErrorUserNotLogin = errors.New("用户未登录")

func GetCurrentUserID(rc *app.RequestContext) (userID uint, err error) {
	uid, ok := rc.Get(ContextUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userID, ok = uid.(uint)
	if !ok {
		err = errors.New("user id type assertion failed")
		return
	}
	return
}
