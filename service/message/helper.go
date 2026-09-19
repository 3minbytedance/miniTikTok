package main

import "go.uber.org/zap"

func errField(err error) zap.Field {
	return zap.Error(err)
}
