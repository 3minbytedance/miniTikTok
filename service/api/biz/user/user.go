package user

import (
	"context"
	"net/http"
	"strconv"

	"douyin/common"
	"douyin/kitex_gen/user"
	"douyin/service/api/rpc"

	"github.com/cloudwego/hertz/pkg/app"
)

func Register(ctx context.Context, c *app.RequestContext) {
	username, password := c.Query("username"), c.Query("password")
	if username == "" || password == "" {
		c.JSON(http.StatusOK, &user.UserRegisterResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.Register(ctx, &user.UserRegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &user.UserRegisterResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func Login(ctx context.Context, c *app.RequestContext) {
	username, password := c.Query("username"), c.Query("password")
	if username == "" || password == "" {
		c.JSON(http.StatusOK, &user.UserLoginResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.Login(ctx, &user.UserLoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &user.UserLoginResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func Info(ctx context.Context, c *app.RequestContext) {
	actorId, _ := common.GetCurrentUserID(c)
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, &user.UserInfoByIdResponse{
			StatusCode: common.CodeInvalidParam,
			StatusMsg:  common.MapErrMsg(common.CodeInvalidParam),
		})
		return
	}
	resp, err := rpc.Social.GetUserInfoById(ctx, &user.UserInfoByIdRequest{
		ActorId: int64(actorId),
		UserId:  userId,
	})
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, &user.UserInfoByIdResponse{
			StatusCode: common.CodeServerBusy,
			StatusMsg:  common.MapErrMsg(common.CodeServerBusy),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}
