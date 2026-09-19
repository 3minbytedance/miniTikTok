package mysql

import (
	"douyin/dal/model"

	"go.uber.org/zap"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// FindUserByName 认证路径（注册判重/登录）强制读主库：
// 主从延迟下避免"刚注册就登录"查不到账号。
func FindUserByName(name string) (user model.User, exist bool, err error) {
	user = model.User{}
	if err = SocialDB.Clauses(dbresolver.Use("source")).Where("name = ?", name).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user, false, nil
		}
		// 处理其他查询错误
		zap.L().Error("Database err", zap.Error(err))
		return user, false, err
	}
	return user, true, nil
}

func FindUserByUserID(id uint) (user model.User, exist bool, err error) {
	user = model.User{}
	if err = SocialDB.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user, false, err
		}
		// 处理其他查询错误
		zap.L().Error("Database err", zap.Error(err))
		return user, false, err
	}
	return user, true, nil
}

// GetUserInfoByID 查询用户资料（头像/背景/简介）。
func GetUserInfoByID(id uint) (info model.UserInfo, err error) {
	err = SocialDB.Where("id = ?", id).First(&info).Error
	return
}

func CreateUser(user *model.User) error {
	userInfo := model.UserInfo{
		ID:   user.ID,
		Name: user.Name,
	}
	tx := SocialDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Create(&userInfo).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
