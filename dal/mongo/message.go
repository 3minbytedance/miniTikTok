package mongo

import (
	"douyin/dal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

const messageCollection = "messages"

// InsertMessage 直接写入一条聊天消息（不再经过 kafka 中转）。
func InsertMessage(message *model.Message) error {
	_, err := Mongo.Collection(messageCollection).InsertOne(Ctx, message)
	if err != nil {
		zap.L().Error("insert message to mongo failed", zap.Error(err))
	}
	return err
}

// GetMessageList 查询 fromUserId 与 toUserId 之间、create_time 晚于 preMsgTime 的消息，按时间正序。
func GetMessageList(fromUserId, toUserId uint, preMsgTime int64) ([]*model.Message, error) {
	filter := bson.M{
		"$and": []bson.M{
			{"$or": []bson.M{
				{"from_user_id": fromUserId, "to_user_id": toUserId},
				{"from_user_id": toUserId, "to_user_id": fromUserId},
			}},
			{"create_time": bson.M{"$gt": preMsgTime}},
		},
	}

	cursor, err := Mongo.Collection(messageCollection).Find(Ctx, filter)
	if err != nil {
		zap.L().Error("query message list failed", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(Ctx)

	var messages []*model.Message
	if err := cursor.All(Ctx, &messages); err != nil {
		zap.L().Error("decode message list failed", zap.Error(err))
		return nil, err
	}
	return messages, nil
}
