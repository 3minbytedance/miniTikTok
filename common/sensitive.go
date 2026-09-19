package common

import (
	"github.com/importcjj/sensitive"
	"go.uber.org/zap"
)

var sensitiveFilter *sensitive.Filter

func InitSensitiveFilter() (err error) {
	sensitiveFilter = sensitive.New()
	err = sensitiveFilter.LoadWordDict("./common/sensitive_word_dic.txt")
	if err != nil {
		zap.L().Error("Load sensitive dic error", zap.Error(err))
		return err
	}
	return
}

func ReplaceWord(word string) string {
	if sensitiveFilter == nil {
		zap.L().Warn("sensitive filter is not initialized")
		return word
	}
	return sensitiveFilter.Replace(word, '*')
}
