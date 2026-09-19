package main

import "go.uber.org/zap"

// errField 统一的错误日志字段。
func errField(err error) zap.Field {
	return zap.Error(err)
}
