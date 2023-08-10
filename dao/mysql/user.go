package mysql

import (
	"fmt"
	"math/rand"
	"project/models"
	"project/utils"
)

func FindUserByName(name string) (models.User, bool) {
	user := models.User{}
	return user, DB.Where("name = ?", name).First(&user).RowsAffected != 0
}

func FindUserStateByName(name string) (models.UserStates, bool) {
	userState := models.UserStates{}
	return userState, DB.Where("name = ?", name).First(&userState).RowsAffected != 0
}

func FindUserByID(id int) (models.User, bool) {
	user := models.User{}
	return user, DB.Where("id = ?", id).First(&user).RowsAffected != 0
}

func FindUserStateByID(id int) (models.UserStates, bool) {
	userState := models.UserStates{}
	return userState, DB.Where("id = ?", id).First(&userState).RowsAffected != 0
}

// FindUserByToken todo 废弃，jwt解析自带信息
func FindUserByToken(token string) (models.User, bool) {
	user := models.User{}
	userState := models.UserStates{}
	row := DB.Where("token = ?", token).First(&userState).RowsAffected
	if row == 0 || userState.IsLogOut {
		return user, false
	}
	// 应该在userStates表里面加id，而不是name
	return user, DB.Where("name = ?", userState.Name).First(&user).RowsAffected != 0
}

func CheckUserRegisterInfo(username string, password string) (int32, string) {

	if len(username) == 0 || len(username) > 32 {
		return 1, "用户名不合法"
	}

	if len(password) <= 6 || len(password) > 32 {
		return 2, "密码不合法"
	}

	if _, ok := FindUserByName(username); ok {
		return 3, "用户已注册"
	}

	return 0, "合法"
}

func RegisterUserInfo(username string, password string) (int32, string, int64, string) {

	// todo 对密码加密
	user := models.User{}
	user.Name = username

	// 生成token，id
	//user.ID = uuid.New()
	// 将信息存储到数据库中

	// salt密码加密
	userStates := models.UserStates{}
	userStates.Name = username
	salt := fmt.Sprintf("%06d", rand.Int())
	userStates.Salt = salt
	userStates.Password = utils.MakePassword(password, salt)
	userStates.Token = utils.GenerateToken(user.ID, username)

	// 数据入库
	DB.Create(&userStates)
	DB.Create(&user)
	fmt.Println("<<<<<<<<<id: ", user.ID)
	userStates.Token = utils.GenerateToken(user.ID, username)
	return 0, "注册成功", int64(user.ID), userStates.Token
}
