package social

import (
	"errors"

	"douyin/dal/model"

	"go.uber.org/zap"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// FindUserByName 认证路径（注册判重/登录）强制读主库：
// 主从延迟下避免"刚注册就登录"查不到账号。
func FindUserByName(name string) (user model.User, exist bool, err error) {
	user = model.User{}
	if err = DB.Clauses(dbresolver.Use("source")).Where("name = ?", name).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, false, nil
		}
		// 处理其他查询错误
		zap.L().Error("Database err", zap.Error(err))
		return user, false, err
	}
	return user, true, nil
}

// FindUserByUserID 按主键查询用户；未找到返回 exist=false 而非错误，与 FindUserByName 语义一致。
func FindUserByUserID(id uint) (user model.User, exist bool, err error) {
	user = model.User{}
	if err = DB.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, false, nil
		}
		zap.L().Error("Database err", zap.Error(err))
		return user, false, err
	}
	return user, true, nil
}

// GetUserInfoByID 查询用户资料（头像/背景/简介）。
func GetUserInfoByID(id uint) (info model.UserInfo, err error) {
	err = DB.Where("id = ?", id).First(&info).Error
	return
}

// CreateUser 同一事务内写入登录表与资料表，任一失败整体回滚。
func CreateUser(user *model.User) error {
	userInfo := model.UserInfo{
		ID:   user.ID,
		Name: user.Name,
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(&userInfo).Error
	})
}
