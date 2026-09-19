package main

import (
	"context"
	"errors"
	"strconv"

	"douyin/common"
	"douyin/dal/model"
	socialdao "douyin/dal/mysql/social"
	"douyin/kitex_gen/user"
	"douyin/mw/redis"
	"douyin/observability"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SocialServiceImpl struct{}

// Register 用户名查重（user_login.name 唯一索引）→ 加密落库 → 发 token。
func (s *SocialServiceImpl) Register(ctx context.Context, req *user.UserRegisterRequest) (*user.UserRegisterResponse, error) {
	resp := &user.UserRegisterResponse{}

	if req.Username == "" {
		resp.StatusCode = common.CodeInvalidRegisterUsername
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidRegisterUsername)
		return resp, nil
	}
	if len(req.Password) < 6 {
		resp.StatusCode = common.CodeInvalidRegisterPassword
		resp.StatusMsg = common.MapErrMsg(common.CodeInvalidRegisterPassword)
		return resp, nil
	}

	// 存在性判断以 MySQL 唯一索引为准（认证路径强制读主库，避免主从延迟）。
	if _, exist, err := socialdao.FindUserByName(req.Username); err != nil {
		observability.Logger(ctx).Error("register query username failed", zap.Error(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	} else if exist {
		resp.StatusCode = common.CodeUsernameAlreadyExists
		resp.StatusMsg = common.MapErrMsg(common.CodeUsernameAlreadyExists)
		return resp, nil
	}

	hashPwd, err := common.MakePassword(req.Password)
	if err != nil {
		observability.Logger(ctx).Error("register make password failed", zap.Error(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}

	uid := common.GetUid()
	newUser := &model.User{ID: uid, Name: req.Username, Password: hashPwd}
	if err := socialdao.CreateUser(newUser); err != nil {
		// 唯一索引冲突说明并发注册同名用户，其余视为服务端异常。
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			resp.StatusCode = common.CodeUsernameAlreadyExists
			resp.StatusMsg = common.MapErrMsg(common.CodeUsernameAlreadyExists)
			return resp, nil
		}
		observability.Logger(ctx).Error("create user failed", zap.Error(err))
		resp.StatusCode = common.CodeServerBusy
		resp.StatusMsg = common.MapErrMsg(common.CodeServerBusy)
		return resp, nil
	}

	SetUserName(ctx, uid, req.Username)

	token := common.GenerateToken(uid, req.Username)
	redis.SetToken(uid, token)

	resp.StatusCode = common.CodeSuccess
	resp.UserId = int64(uid)
	resp.Token = token
	return resp, nil
}

// Login 查库校验密码 → 发 token。
func (s *SocialServiceImpl) Login(ctx context.Context, req *user.UserLoginRequest) (*user.UserLoginResponse, error) {
	resp := &user.UserLoginResponse{}

	dbUser, exist, err := socialdao.FindUserByName(req.Username)
	if err != nil || !exist {
		resp.StatusCode = common.CodeUsernameNotFound
		resp.StatusMsg = common.MapErrMsg(common.CodeUsernameNotFound)
		return resp, nil
	}
	if !common.CheckPassword(req.Password, dbUser.Password) {
		resp.StatusCode = common.CodeWrongLoginCredentials
		resp.StatusMsg = common.MapErrMsg(common.CodeWrongLoginCredentials)
		return resp, nil
	}

	token := common.GenerateToken(dbUser.ID, dbUser.Name)
	redis.SetToken(dbUser.ID, token)

	resp.StatusCode = common.CodeSuccess
	resp.UserId = int64(dbUser.ID)
	resp.Token = token
	return resp, nil
}

// GetUserInfoById 组装用户主页信息：关系计数本地查，作品/点赞计数 RPC 查 video 域。
func (s *SocialServiceImpl) GetUserInfoById(ctx context.Context, req *user.UserInfoByIdRequest) (*user.UserInfoByIdResponse, error) {
	resp := &user.UserInfoByIdResponse{}
	u, err := s.buildUser(ctx, uint(req.UserId), uint(req.ActorId))
	if err != nil {
		resp.StatusCode = common.CodeUserNotFound
		resp.StatusMsg = common.MapErrMsg(common.CodeUserNotFound)
		return resp, nil
	}
	resp.StatusCode = common.CodeSuccess
	resp.User = u
	return resp, nil
}

// buildUser 组装完整用户信息。actorId 为当前访问者，用于计算 is_follow。
func (s *SocialServiceImpl) buildUser(ctx context.Context, uid, actorId uint) (*user.User, error) {
	u := &user.User{Id: int64(uid)}

	name, err := GetUserName(ctx, uid)
	if err != nil {
		return nil, err
	}
	u.Name = name

	if info, err := socialdao.GetUserInfoByID(uid); err == nil {
		u.Avatar = info.Avatar
		u.BackgroundImage = info.BackgroundImage
		u.Signature = info.Signature
	}

	if fc, err := GetFollowCount(ctx, uid); err == nil {
		u.FollowCount = fc
	}
	if frc, err := GetFollowerCount(ctx, uid); err == nil {
		u.FollowerCount = frc
	}

	// video 域统计（跨域 RPC，计数有短 TTL 缓存）。
	if videoClient != nil {
		if wc, err := videoClient.GetWorkCount(ctx, int64(uid)); err == nil {
			u.WorkCount = wc
		}
		if fc, err := videoClient.GetUserFavoriteCount(ctx, int64(uid)); err == nil {
			u.FavoriteCount = fc
		}
		if tf, err := videoClient.GetUserTotalFavoritedCount(ctx, int64(uid)); err == nil {
			u.TotalFavorited = strconv.FormatInt(int64(tf), 10)
		}
	}

	if actorId != 0 && actorId != uid {
		if followed, err := IsFollowing(ctx, actorId, uid); err == nil {
			u.IsFollow = followed
		}
	}
	return u, nil
}
