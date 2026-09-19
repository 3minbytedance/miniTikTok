# 所有构建产物统一输出到 build/ 目录
BIN_DIR  := build/bin

# 让 make 能找到通过 `go install` 安装的 kitex / hz
export PATH := $(shell go env GOPATH)/bin:$(PATH)
GOPROXY ?= https://goproxy.cn,direct

.PHONY: all gen build-api build-social build-video build-message build-all tidy fmt vet clean

all: build-all

## gen: 根据 idl/ 重新生成 kitex_gen
gen:
	kitex -module douyin -I idl idl/social.thrift
	kitex -module douyin -I idl idl/videoapp.thrift
	kitex -module douyin -I idl idl/message.thrift

## 单独构建某个服务，产物输出到 build/bin/
build-api build-social build-video build-message:
	@mkdir -p $(BIN_DIR)
	GOPROXY=$(GOPROXY) go build -o $(BIN_DIR)/$(patsubst build-%,%,$@) ./service/$(patsubst build-%,%,$@)

## 构建全部服务
build-all: build-api build-social build-video build-message

## tidy: 整理依赖
tidy:
	GOPROXY=$(GOPROXY) go mod tidy

## fmt: 格式化代码
fmt:
	gofmt -w .

## vet: 静态检查
vet:
	GOPROXY=$(GOPROXY) go vet ./...

## clean: 清理构建产物
clean:
	rm -rf build/
