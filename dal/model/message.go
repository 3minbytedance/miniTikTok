package model

type Message struct {
	ID         int64  `bson:"id" json:"id"`
	FromUserId uint   `bson:"from_user_id" json:"from_user_id"`
	ToUserId   uint   `bson:"to_user_id" json:"to_user_id"`
	Content    string `bson:"content" json:"content,omitempty"`
	CreateTime int64  `bson:"create_time" json:"create_time"`
}

type MessageChatResponse struct {
	Response
	MessageList []Message `json:"message_list"`
}

type MessageSendEvent struct {
	UserId     int64  `json:"user_id"`
	ToUserId   int64  `json:"to_user_id"`
	MsgContent string `json:"msg_content"`
}

type MessagePushEvent struct {
	FromUserId int64  `json:"user_id"`
	MsgContent string `json:"msg_content"`
}
