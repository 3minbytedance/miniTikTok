package model

type UserFollow struct {
	// 用户的关注信息
	ID uint `gorm:"primaryKey"`
	// (user_id, follow_id) 联合唯一：既防重复关注，也能覆盖"查我的关注/关注数"的 user_id 前缀查询
	UserId   uint `gorm:"uniqueIndex:uk_user_follow,priority:1;not null"`                     // 用户id
	FollowId uint `gorm:"uniqueIndex:uk_user_follow,priority:2;index:idx_follow_id;not null"` // 关注用户id
}

func (*UserFollow) TableName() string {
	return "user_follow"
}
