package model

type Favorite struct {
	ID uint `gorm:"primaryKey"`
	// (user_id, video_id) 联合唯一：既防重复点赞，也能覆盖"我的点赞数/点赞列表"的 user_id 前缀查询
	UserId  uint `gorm:"uniqueIndex:uk_user_video,priority:1;not null"`
	VideoId uint `gorm:"uniqueIndex:uk_user_video,priority:2;index:idx_video_id;not null"` // 支撑按视频统计点赞数
	// 点赞时间（秒），作为点赞列表 ZSet 的 score；列表排序仍用自增 id
	CreatedAt int64 `gorm:"autoCreateTime"`
}

type FavoriteAction struct {
	UserId     uint
	VideoId    uint
	ActionType int
}

func (*Favorite) TableName() string {
	return "favorite"
}

type FavoriteListResponse struct {
	FavoriteRes   Response
	VideoResponse []VideoResponse `json:"video_list"`
}
