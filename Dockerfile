# syntax=docker/dockerfile:1
# 多服务共用一个 Dockerfile，通过构建参数 SERVICE 区分：
#   docker build --build-arg SERVICE=social -t douyin-social .
# SERVICE 取值：api | social | video | message

########## stage 1: build ##########
FROM golang:1.26-bookworm AS builder

ARG SERVICE=api
# 国内构建默认走 goproxy.cn，可在构建时覆盖：--build-arg GOPROXY=https://proxy.golang.org,direct
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY} \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

# 先拉依赖，利用分层缓存
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# 只编译目标服务
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/bin/${SERVICE} ./service/${SERVICE}

########## stage 2: runtime ##########
FROM debian:bookworm-slim

ARG SERVICE=api
# 国内构建默认用 USTC Debian 镜像源，可覆盖：--build-arg APT_MIRROR=deb.debian.org
ARG APT_MIRROR=mirrors.ustc.edu.cn
ENV TZ=Asia/Shanghai

# ca-certificates：HTTPS/OTLP 需要；ffmpeg：OSS 未启用时本地视频截帧依赖
RUN sed -i "s|deb.debian.org|${APT_MIRROR}|g" /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 二进制、配置文件、敏感词词典（运行时按相对路径加载）
COPY --from=builder /out/bin/${SERVICE} /app/bin/service
COPY --from=builder /build/config ./config
COPY --from=builder /build/common/sensitive_word_dic.txt ./common/sensitive_word_dic.txt

RUN mkdir -p /app/tmp /app/log

# api=8080 social=4002 video=4003 message=4005
EXPOSE 8080 4002 4003 4005

ENTRYPOINT ["/app/bin/service"]
