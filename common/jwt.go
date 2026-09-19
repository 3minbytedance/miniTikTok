package common

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Claims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 签名密钥
var jwtSecretKey = []byte("DouShenNo1")

// GenerateToken 创建token， 过期时间设置已注释
func GenerateToken(userId uint, username string) string {

	nowTime := time.Now()
	//expireTime := nowTime.Add(24 * time.Hour).Unix()
	//log.Println("expireTime:", expireTime)

	claims := Claims{
		ID:       userId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(nowTime),
			Issuer:   "DouShen",
		},
	}
	// 使用用于签名的算法和令牌
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 创建JWT字符串
	if token, err := tokenClaims.SignedString(jwtSecretKey); err != nil {
		zap.L().Error("generate token fail!", zap.Error(err))
		return "fail"
	} else {
		zap.L().Info("generate token success!")
		return token
	}
}

// ParseToken 解析token
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})
	if err != nil {
		return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
